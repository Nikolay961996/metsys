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

	r.Get("/", WithLogger(getDashboardHandler(s)))
	r.Get("/value/{metricType}/{metricName}", WithLogger(getMetricValueHandler(s)))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", WithLogger(updateMetricHandler(s)))
	r.Post("/value/", WithLogger(getMetricValueJSONHandler(s)))
	r.Post("/update/", WithLogger(updateMetricJSONHandler(s)))

	r.Post("/update/*", WithLogger(updateErrorPathHandler()))

	return r
}
