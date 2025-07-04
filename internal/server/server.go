package server

import (
	"database/sql"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
)

type MetricServer struct {
	Storage repositories.Storage
	DB      *sql.DB

	config *Config
}

func InitServer(c *Config) MetricServer {
	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		panic(err)
	}

	var s repositories.Storage

	if c.DatabaseDSN != "" {
		s = storage.NewDBStorage(c.DatabaseDSN)
	} else if c.FileStoragePath != "" {
		s = storage.NewFileStorage(c.FileStoragePath, c.StoreInterval, c.Restore) // c.FileStoragePath, c.StoreInterval == 0, c.Restore
	} else {
		s = storage.NewMemStorage()
	}

	a := MetricServer{
		DB:      db,
		config:  c,
		Storage: s,
	}
	return a
}

func (s *MetricServer) Run() {
	err := http.ListenAndServe(s.config.RunOnServerAddress, router.MetricsRouterWithServer(s.Storage, s.DB))
	if err != nil {
		models.Log.Error(err.Error())
	}
}

func (s *MetricServer) Stop() {
	models.Log.Warn("Server shutting down")
	_ = s.DB.Close()
	s.Storage.Close()
}
