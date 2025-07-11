package router

import (
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

func MetricsRouter() *chi.Mux {
	s := storage.NewMemStorage("/local.db", false, false)
	return MetricsRouterWithServer(s)
}

func MetricsRouterWithServer(s *storage.MemStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(WithDecompressionRequest, WithLogger)

	r.Get("/", WithCompressionResponse(getDashboardHandler(s)))
	r.Get("/value/{metricType}/{metricName}", getMetricValueHandler(s))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetricHandler(s))
	r.Post("/value/", WithCompressionResponse(getMetricValueJSONHandler(s)))
	r.Post("/update/", WithCompressionResponse(updateMetricJSONHandler(s)))

	r.Post("/update/*", updateErrorPathHandler())

	return r
}
