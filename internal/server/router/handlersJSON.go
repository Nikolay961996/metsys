// Package router consist json handlers
package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
)

func isCorrectMethod(expectedMethod string, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Only %s method allowed", expectedMethod), http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func baseJSONHandler(w http.ResponseWriter, r *http.Request, storage repositories.Storage, innerFunc func(http.ResponseWriter, repositories.Storage, *models.Metrics) bool) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	if !isCorrectMethod(http.MethodPost, w, r) {
		return
	}
	mr := readJSONMetrics(w, r)
	if mr == nil {
		return
	}

	if innerFunc != nil {
		ok := innerFunc(w, storage, mr)
		if !ok {
			return
		}
	}

	actualMr, err := GetActualMetrics(storage, mr)
	if err != nil {
		models.Log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSONMetrics(w, actualMr)
}

func getMetricValueJSONHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseJSONHandler(w, r, storage, nil)
	}
}

func updateMetricJSONHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseJSONHandler(w, r, storage, updateMetrics)
	}
}

func updatesMetricJSONHandler(storage repositories.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models.Log.Info("Get batch metrics")
		w.Header().Set("content-type", "application/json; charset=utf-8")
		if !isCorrectMethod(http.MethodPost, w, r) {
			return
		}
		mrs := readJSONMetricsArray(w, r)
		if mrs == nil {
			return
		}
		models.Log.Info(fmt.Sprintf("Batch: %v", mrs))

		ctx := context.Background()
		err := storage.StartTransaction(ctx)
		if err != nil {
			models.Log.Error(fmt.Sprintf("Error start transaction: %v", err))
			http.Error(w, fmt.Sprintf("Error start transaction: %v", err), http.StatusBadRequest)
			return
		}
		for _, mr := range mrs {
			_ = updateMetrics(w, storage, &mr)
		}

		err = storage.CommitTransaction()

		if err != nil {
			models.Log.Error(fmt.Sprintf("Error commit transaction: %v", err))
			http.Error(w, fmt.Sprintf("Error commit transaction: %v", err), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetActualMetrics(storage repositories.Storage, mr *models.Metrics) (*models.Metrics, error) {
	var actual = models.Metrics{
		ID:    mr.ID,
		MType: mr.MType,
		Value: nil,
		Delta: nil,
	}

	switch mr.MType {
	case models.Gauge:
		v, err := storage.GetGauge(mr.ID)
		if err != nil {
			return nil, errors.New("metric not found")
		}
		actual.Value = &v
	case models.Counter:
		v, err := storage.GetCounter(mr.ID)
		if err != nil {
			return nil, errors.New("metric not found")
		}
		actual.Delta = &v
	default:
		return nil, errors.New("metric type not found")
	}

	return &actual, nil
}

func updateMetrics(w http.ResponseWriter, storage repositories.Storage, mr *models.Metrics) bool {
	if mr.MType == models.Gauge {
		storage.SetGauge(mr.ID, *mr.Value)
	} else if mr.MType == models.Counter {
		storage.AddCounter(mr.ID, *mr.Delta)
	} else {
		models.Log.Error(fmt.Sprintf("Error undefind type: %v", mr.MType))
		http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", mr.MType), http.StatusBadRequest)
		return false
	}

	return true
}

func writeJSONMetrics(w http.ResponseWriter, metrics *models.Metrics) {
	resp, err := json.Marshal(metrics)
	if err != nil {
		models.Log.Error(fmt.Sprintf("Error marshalling body: %v", err))
		http.Error(w, fmt.Sprintf("Error marshalling body: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		models.Log.Error(fmt.Sprintf("Error writing response: %v", err))
	}
}

func readJSONMetrics(w http.ResponseWriter, r *http.Request) *models.Metrics {
	var mr models.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()

	if err != nil {
		models.Log.Error(fmt.Sprintf("Error reading body: %v", err))
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
		return nil
	}

	if err := json.Unmarshal(buf.Bytes(), &mr); err != nil {
		models.Log.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", err), http.StatusBadRequest)
		return nil
	}

	return &mr
}

func readJSONMetricsArray(w http.ResponseWriter, r *http.Request) []models.Metrics {
	var mr []models.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()

	if err != nil {
		models.Log.Error(fmt.Sprintf("Error reading body: %v", err))
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
		return nil
	}

	if err := json.Unmarshal(buf.Bytes(), &mr); err != nil {
		models.Log.Error(fmt.Sprintf("Error unmarshalling body: %v", err))
		http.Error(w, fmt.Sprintf("Error unmarshalling body: %v", err), http.StatusBadRequest)
		return nil
	}

	return mr
}
