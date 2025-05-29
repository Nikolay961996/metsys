package main

import (
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

type Storage interface {
	SetGauge(metricName string, value float64)
	GetGauge(metricName string) (float64, error)
	AddCounter(metricName string, value int64)
	GetCounter(metricName string) (int64, error)
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.GaugeMetrics[metricName] = value
}

func (m *MemStorage) GetGauge(metricName string) (float64, error) {
	value, ok := m.GaugeMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *MemStorage) AddCounter(metricName string, value int64) {
	m.CounterMetrics[metricName] += value
}

func (m *MemStorage) GetCounter(metricName string) (int64, error) {
	value, ok := m.CounterMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func newMemStorage() *MemStorage {
	return &MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	storage := newMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) { updateMetricHandler(w, r, storage) })

	return http.ListenAndServe(":8080", mux)
}

// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func updateMetricHandler(w http.ResponseWriter, r *http.Request, storage Storage) {
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
