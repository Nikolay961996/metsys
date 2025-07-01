package server

import (
	"database/sql"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
	"time"
)

type MetricServer struct {
	Storage *storage.MemStorage
	DB      *sql.DB

	config     *Config
	isSyncSave bool
	saveTimer  *time.Ticker
}

func InitServer(c *Config) MetricServer {
	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	a := MetricServer{
		DB:         db,
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
	err := http.ListenAndServe(s.config.RunOnServerAddress, router.MetricsRouterWithServer(s.Storage, s.DB))
	if err != nil {
		models.Log.Error(err.Error())
	}
}

func (s *MetricServer) Stop() {
	models.Log.Warn("Server shutting down")
	_ = s.DB.Close()
	s.saveTimer.Stop()
}

func (s *MetricServer) backgroundSaver() {
	for range s.saveTimer.C {
		s.Storage.TryFlushToFile()
	}
}
