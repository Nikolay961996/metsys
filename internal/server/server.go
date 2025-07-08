package server

import (
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
)

type MetricServer struct {
	Storage repositories.Storage
}

func InitServer(c *Config) MetricServer {
	a := MetricServer{}

	if c.DatabaseDSN != "" {
		a.Storage = storage.NewDBStorage(c.DatabaseDSN)
	} else if c.FileStoragePath != "" {
		a.Storage = storage.NewFileStorage(c.FileStoragePath, c.StoreInterval, c.Restore) // c.FileStoragePath, c.StoreInterval == 0, c.Restore
	} else {
		a.Storage = storage.NewMemStorage()
	}

	return a
}

func (s *MetricServer) Run(runOnServerAddress string) {
	err := http.ListenAndServe(runOnServerAddress, router.MetricsRouterWithServer(s.Storage))
	if err != nil {
		models.Log.Error(err.Error())
	}
}

func (s *MetricServer) Stop() {
	models.Log.Warn("Server shutting down")
	s.Storage.Close()
}
