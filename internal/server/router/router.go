package router

import (
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

func MetricsRouter() *chi.Mux {
	metricsStorage := storage.NewMemStorage()
	return MetricsRouterWithStorage(metricsStorage)
}

func MetricsRouterWithStorage(s *storage.MemStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(WithCompression, WithCompressionResponse, WithLogger)

	r.Get("/", getDashboardHandler(s))
	r.Get("/value/{metricType}/{metricName}", getMetricValueHandler(s))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetricHandler(s))
	r.Post("/value/", getMetricValueJSONHandler(s))
	r.Post("/update/", updateMetricJSONHandler(s))

	r.Post("/update/*", updateErrorPathHandler())

	return r
}
