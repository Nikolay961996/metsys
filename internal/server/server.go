// Package server consist server main entities
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"github.com/Nikolay961996/metsys/proto"
	"google.golang.org/grpc"
)

type MetricServer struct {
	Storage repositories.Storage
	srv     *http.Server
	grpcSrv *grpc.Server // Added gRPC server instance
}

func InitServer(c *Config) MetricServer {
	a := MetricServer{}

	if c.DatabaseDSN != "" {
		a.Storage = storage.NewDBStorage(c.DatabaseDSN)
	} else if c.FileStoragePath != "" {
		a.Storage = storage.NewFileStorage(c.FileStoragePath, c.StoreInterval, c.Restore)
	} else {
		a.Storage = storage.NewMemStorage()
	}

	return a
}

func (s *MetricServer) Run(c *Config) {
	if c.GRPCPort == "" && c.RunOnServerAddress == "" {
		panic("No port specified for either HTTP or gRPC server")
	}

	if c.GRPCPort != "" {
		s.RunGRPC(c.GRPCPort)
	}

	if c.RunOnServerAddress != "" {
		privateKey, err := crypto.ParseRSAPrivateKeyPEM(c.CryptoKey)
		if err != nil {
			panic(fmt.Errorf("error parsing private key: %v", err))
		}

		handler := router.MetricsRouterWithServer(s.Storage, c.KeyForSigning, privateKey, c.TrustedSubnet)
		s.srv = &http.Server{
			Addr:    c.RunOnServerAddress,
			Handler: handler,
		}

		runBackground(s)
	}
}

func (s *MetricServer) RunGRPC(grpcPort string) {
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		panic(fmt.Errorf("failed to listen on gRPC port %s: %v", grpcPort, err))
	}

	s.grpcSrv = grpc.NewServer()

	proto.RegisterMetricsServiceServer(s.grpcSrv, &MetricsServiceServer{Storage: s.Storage})

	go func() {
		if err := s.grpcSrv.Serve(listener); err != nil {
			panic(fmt.Errorf("failed to serve gRPC: %v", err))
		}
	}()
}

// Stop gracefully shuts down the HTTP server and closes storage
func (s *MetricServer) Stop(timeout time.Duration) {
	models.Log.Warn("Server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if s.srv != nil {
		if err := s.srv.Shutdown(ctx); err != nil {
			models.Log.Error("server shutdown error: " + err.Error())
		}
	}
	s.Storage.Close()
}

func runBackground(s *MetricServer) {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			models.Log.Error("listen error: " + err.Error())
		}
	}()
}
