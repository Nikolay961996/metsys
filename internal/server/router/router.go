// Package router consist register routes with handlers
package router

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/storage"
)

func MetricsRouterTest() *chi.Mux {
	s := storage.NewFileStorage("/local.db", 5*time.Second, false)

	return MetricsRouterWithServer(s, "")
}

func MetricsRouterWithServer(s repositories.Storage, keyForSigning string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		WithDecompressionRequest,
		WithLogger,
		WithSigningCheck(keyForSigning),
		WithSigningResponse(keyForSigning))

	r.Get("/", WithCompressionResponse(getDashboardHandler(s)))
	r.Get("/ping", pingDatabase(s))

	r.Get("/value/{metricType}/{metricName}", getMetricValueHandler(s))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateMetricHandler(s))

	r.Post("/value/", WithCompressionResponse(getMetricValueJSONHandler(s)))
	r.Post("/update/", WithCompressionResponse(updateMetricJSONHandler(s)))

	r.Post("/updates/", WithCompressionResponse(updatesMetricJSONHandler(s)))

	r.Post("/update/*", updateErrorPathHandler())

	return r
}
