// Package agent internal logic.
package agent

import (
	"context"
	"errors"
	"time"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/models"
)

// Entity for agent
type Entity struct {
	doneCtx context.Context    // for cancel
	cancel  context.CancelFunc // for call cancel
}

// InitAgent creating new agent entity
func InitAgent() Entity {
	ctx, c := context.WithCancel(context.Background())
	a := Entity{
		doneCtx: ctx,
		cancel:  c,
	}

	return a
}

// Run agent
func (a *Entity) Run(config *Config) {
	publicKey, err := crypto.ParseRSAPublicKeyPEM(config.CryptoKey)
	if err != nil {
		panic(errors.New("parse RSA public key failed"))
	}

	jobsChan := make(chan workerJob, config.SendMetricsRateLimit)
	newMetricsChan := runPollWorker(config.PollInterval, a.doneCtx)
	newGopsutilMetricsChan := runPollGopsutilWorker(config.PollInterval, a.doneCtx)
	for i := 0; i < config.SendMetricsRateLimit; i++ {
		go runReportWorker(i, a.doneCtx, jobsChan, config.SendToServerAddress, config.KeyForSigning, publicKey)
	}

	listenMetricsAndFadeOut(a.doneCtx, config.ReportInterval, newMetricsChan, newGopsutilMetricsChan, jobsChan)
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
