package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func getDashboardHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := storage.GetAll()

		t, err := template.ParseFiles("./internal/server/router/metrics.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := t.Execute(w, metrics); err != nil {
			http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
			return
		}
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

func getMetricValueJSONHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricsReq models.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		defer r.Body.Close()

		if err != nil {
			models.Log.Error(fmt.Sprintf("Error reading body: %v", err))
			http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &metricsReq); err != nil {
			models.Log.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
			http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", err), http.StatusBadRequest)
			return
		}

		switch metricsReq.MType {
		case models.Gauge:
			v, err := storage.GetGauge(metricsReq.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metricsReq.Value = &v
		case models.Counter:
			v, err := storage.GetCounter(metricsReq.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metricsReq.Delta = &v
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(metricsReq)
		if err != nil {
			models.Log.Error(fmt.Sprintf("Error marshalling body: %v", err))
			http.Error(w, fmt.Sprintf("Error marshalling body: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			models.Log.Error(fmt.Sprintf("Error writing response: %v", err))
		}
	}
}

func updateMetricHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		/*
			if r.Header.Get("Content-Type") != "text/plain" {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
		*/

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

func updateMetricJSONHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		var mr models.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		defer r.Body.Close()

		if err != nil {
			models.Log.Error(fmt.Sprintf("Error reading body: %v", err))
			http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &mr); err != nil {
			models.Log.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
			http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", err), http.StatusBadRequest)
			return
		}

		if mr.MType == models.Gauge {
			storage.SetGauge(mr.ID, *mr.Value)
			v, _ := storage.GetGauge(mr.ID)
			mr.Value = &v
		} else if mr.MType == models.Counter {
			storage.AddCounter(mr.ID, *mr.Delta)
			v, _ := storage.GetCounter(mr.ID)
			mr.Delta = &v
		} else {
			models.Log.Error(fmt.Sprintf("Error undefind type: %v", mr.MType))
			http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", mr.MType), http.StatusBadRequest)
		}

		resp, err := json.Marshal(mr)
		if err != nil {
			models.Log.Error(fmt.Sprintf("Error marshalling body: %v", err))
			http.Error(w, fmt.Sprintf("Error marshalling body: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			models.Log.Error(fmt.Sprintf("Error writing response: %v", err))
		}
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
