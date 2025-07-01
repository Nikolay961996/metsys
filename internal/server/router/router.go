package router

import (
	"database/sql"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

func MetricsRouterTest() *chi.Mux {
	s := storage.NewMemStorage("/local.db", false, false)
	db, err := sql.Open("pgx", "host=localhost user=postgres password=admin dbname=metsys sslmode=disable")
	if err != nil {
		panic(err)
	}

	return MetricsRouterWithServer(s, db)
}

func MetricsRouterWithServer(s *storage.MemStorage, db *sql.DB) *chi.Mux {
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
