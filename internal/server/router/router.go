package router

import (
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func MetricsRouter() *chi.Mux {
	metricsStorage := storage.NewMemStorage()
	return MetricsRouterWithStorage(metricsStorage)
}

func MetricsRouterWithStorage(s *storage.MemStorage) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/value/{metricType}/{metricName}", getMetricValueHandler(s))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetricHandler(s))

	r.Post("/update/*", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 4 {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}

		if parts[1] != models.Gauge && parts[1] != models.Counter {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		http.NotFound(w, r)
	})

	return r
}
