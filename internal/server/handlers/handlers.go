package handlers

import (
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
	"strconv"
	"strings"
)

// UpdateMetricHandler http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func UpdateMetricHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("content-type", "text/plain; charset=utf-8")
		metricName, metricType, counterValue, gaugeValue, err := parseMetricData(r.URL.Path, w)
		fmt.Println(metricName, counterValue, gaugeValue, err)
		if err != nil {
			return
		}

		if metricType == models.Gauge {
			storage.SetGauge(metricName, gaugeValue)
		} else if metricType == models.Counter {
			storage.AddCounter(metricName, counterValue)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func parseMetricData(url string, w http.ResponseWriter) (string, string, int64, float64, error) {
	path := strings.TrimPrefix(url, "/update/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid format", http.StatusNotFound)
		return "", "", 0, 0, errors.New("invalid format")
	}

	metricType := parts[0]
	metricName := parts[1]
	metricValueStr := parts[2]

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
