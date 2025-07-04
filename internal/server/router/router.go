package router

import (
	"database/sql"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"time"
)

func MetricsRouterTest() *chi.Mux {
	s := storage.NewFileStorage("/local.db", 5*time.Second, false)
	db, err := sql.Open("pgx", "host=localhost user=postgres password=admin dbname=metsys sslmode=disable")
	if err != nil {
		panic(err)
	}

	return MetricsRouterWithServer(s, db)
}

func MetricsRouterWithServer(s repositories.Storage, db *sql.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(WithDecompressionRequest, WithLogger)

	r.Get("/", WithCompressionResponse(getDashboardHandler(s)))
	r.Get("/ping", pingDatabase(db))

	r.Get("/value/{metricType}/{metricName}", getMetricValueHandler(s))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetricHandler(s))

	r.Post("/value/", WithCompressionResponse(getMetricValueJSONHandler(s)))
	r.Post("/update/", WithCompressionResponse(updateMetricJSONHandler(s)))

	r.Post("/update/*", updateErrorPathHandler())

	return r
}
