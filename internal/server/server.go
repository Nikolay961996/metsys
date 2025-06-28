package server

import (
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
	"time"
)

type MetricServer struct {
	Storage *storage.MemStorage

	config     *Config
	isSyncSave bool
	saveTimer  *time.Ticker
}

func InitServer(c *Config) MetricServer {
	a := MetricServer{
		config:     c,
		Storage:    storage.NewMemStorage(c.FileStoragePath, c.StoreInterval == 0, c.Restore),
		isSyncSave: c.StoreInterval == 0,
		saveTimer:  time.NewTicker(c.StoreInterval),
	}
	return a
}

func (s *MetricServer) Run() {
	if !s.isSyncSave {
		go s.backgroundSaver()
	}
	err := http.ListenAndServe(s.config.RunOnServerAddress, router.MetricsRouterWithServer(s.Storage))
	if err != nil {
		models.Log.Error(err.Error())
	}
}

func (s *MetricServer) Stop() {
	models.Log.Warn("Server shutting down")
	s.saveTimer.Stop()
}

func (s *MetricServer) backgroundSaver() {
	for range s.saveTimer.C {
		s.Storage.TryFlushToFile()
	}
}
