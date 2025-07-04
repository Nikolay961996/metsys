package router

import (
	"context"
	"errors"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func pingDatabase(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := storage.PingContext(ctx); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getMetricValueHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		var result string
		switch metricType {
		case models.Gauge:
			v, err := storage.GetGauge(metricName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			result = strconv.FormatFloat(v, 'f', -1, 64)
		case models.Counter:
			v, err := storage.GetCounter(metricName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			result = strconv.FormatInt(v, 10)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err := io.WriteString(w, result)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func updateMetricHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		metricName, metricType, counterValue, gaugeValue, err := parseMetricData(r, w)
		if err != nil {
			return
		}
		if metricType == models.Gauge {
			storage.SetGauge(metricName, gaugeValue)
		} else if metricType == models.Counter {
			storage.AddCounter(metricName, counterValue)
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

func updateErrorPathHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func parseMetricData(r *http.Request, w http.ResponseWriter) (string, string, int64, float64, error) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValueStr := chi.URLParam(r, "metricValue")

	if len(metricName) == 0 {
		http.Error(w, "Metric name is empty", http.StatusNotFound)
		return "", "", 0, 0, errors.New("metric name is empty")
	}

	if metricType != models.Counter && metricType != models.Gauge {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return "", "", 0, 0, errors.New("invalid metric type")
	}

	var counterValue int64
	var gaugeValue float64
	var err error
	if metricType == models.Counter {
		counterValue, err = strconv.ParseInt(metricValueStr, 10, 64)
	} else {
		gaugeValue, err = strconv.ParseFloat(metricValueStr, 64)
	}

	if err != nil {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return "", "", 0, 0, errors.New("invalid metric value")
	}

	return metricName, metricType, counterValue, gaugeValue, nil
}
