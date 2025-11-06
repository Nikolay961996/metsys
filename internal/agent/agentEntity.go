// Package agent internal logic.
package agent

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Entity for agent
type Entity struct {
	doneCtx  context.Context    // for cancel
	cancel   context.CancelFunc // for call cancel
	jobsChan chan workerJob
	wg       sync.WaitGroup
}

// InitAgent creating new agent entity
func InitAgent() *Entity {
	ctx, c := context.WithCancel(context.Background())
	a := Entity{
		doneCtx: ctx,
		cancel:  c,
	}

	return &a
}

// Run agent
func (a *Entity) Run(config *Config) {
	if config.GRPCServerAddress == "" && config.SendToServerAddress == "" {
		panic("No server address specified for either HTTP or gRPC communication")
	}

	if config.GRPCServerAddress != "" {
		go a.RunGRPC(config)
	}

	if config.SendToServerAddress != "" {
		publicKey, err := crypto.ParseRSAPublicKeyPEM(config.CryptoKey)
		if err != nil {
			panic(errors.New("parse RSA public key failed"))
		}

		realIP := getRealIP()
		jobsChan := make(chan workerJob, config.SendMetricsRateLimit)
		a.jobsChan = jobsChan

		newMetricsChan := runPollWorker(config.PollInterval, a.doneCtx)
		newGopsutilMetricsChan := runPollGopsutilWorker(config.PollInterval, a.doneCtx)

		for i := 0; i < config.SendMetricsRateLimit; i++ {
			a.wg.Add(1)
			go func(id int) {
				defer a.wg.Done()
				runReportWorker(id, jobsChan, config.SendToServerAddress, config.KeyForSigning, publicKey, realIP)
			}(i)
		}

		listenMetricsAndFadeOut(a.doneCtx, config.ReportInterval, newMetricsChan, newGopsutilMetricsChan, jobsChan)

		a.wg.Wait()
	}
}

// RunGRPC runs gRPC client for communication with the server
func (a *Entity) RunGRPC(config *Config) {
	clientConn, err := grpc.NewClient(config.GRPCServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(errors.New("failed to connect to gRPC server: " + err.Error()))
	}
	defer clientConn.Close()

	// Example: Initialize gRPC client and perform operations
	// client := pb.NewYourServiceClient(clientConn)
	// Call gRPC methods using the client

	// Placeholder for gRPC communication logic
	models.Log.Info("Connected to gRPC server at " + config.GRPCServerAddress)
}

// Stop agent
func (a *Entity) Stop() {
	a.cancel()
}

func listenMetricsAndFadeOut(doneCtx context.Context, period time.Duration, metricsCn <-chan Metrics, gopsutilMetricsCn <-chan MetricsGopsutil, jobsChan chan<- workerJob) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	var metrics Metrics
	var gopsutilMetrics MetricsGopsutil

	defer func() {
		metricsArray := createMetricsArray(&metrics)
		for _, m := range metricsArray {
			jobsChan <- workerJob{oneMetrics: m}
		}
		gMetricsArray := createGopsutilMetricsArray(&gopsutilMetrics)
		for _, m := range gMetricsArray {
			jobsChan <- workerJob{oneMetrics: m}
		}
		close(jobsChan)
	}()

	for {
		select {
		case newMetrics := <-metricsCn:
			metrics = newMetrics
		case newGMetrics := <-gopsutilMetricsCn:
			gopsutilMetrics = newGMetrics
		case <-ticker.C:
			metricsArray := createMetricsArray(&metrics)
			for _, m := range metricsArray {
				jobsChan <- workerJob{oneMetrics: m}
			}
			metrics.PollCount = 0

			gMetricsArray := createGopsutilMetricsArray(&gopsutilMetrics)
			for _, m := range gMetricsArray {
				jobsChan <- workerJob{oneMetrics: m}
			}
		case <-doneCtx.Done():
			models.Log.Warn("Listen fadeOut closed")
			return
		}
	}
}

func getRealIP() string {
	realIP := "127.0.0.1" // Default
	ifaces, err := net.Interfaces()
	if err != nil {
		models.Log.Error("Failed to get network interfaces: " + err.Error())
	} else {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				models.Log.Error("Failed to get addresses for interface: " + iface.Name)
				continue
			}

			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					if v.IP.To4() != nil {
						realIP = v.IP.String()
						break
					}
				case *net.IPAddr:
					if v.IP.To4() != nil {
						realIP = v.IP.String()
						break
					}
				}
			}
		}
	}

	return realIP
}
