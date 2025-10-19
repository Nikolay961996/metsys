// Package server consist server main entities
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
)

type MetricServer struct {
	Storage repositories.Storage
	srv     *http.Server
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

func (s *MetricServer) Run(runOnServerAddress string, keyForSigning string, cryptoKey string) {
	privateKey, err := crypto.ParseRSAPrivateKeyPEM(cryptoKey)
	if err != nil {
		panic(fmt.Errorf("error parsing private key: %v", err))
	}

	handler := router.MetricsRouterWithServer(s.Storage, keyForSigning, privateKey)
	s.srv = &http.Server{
		Addr:    runOnServerAddress,
		Handler: handler,
	}

	runBackground(s)
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
