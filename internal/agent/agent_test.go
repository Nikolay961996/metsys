package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsCollect(t *testing.T) {
	var metrics Metrics
	Poll(&metrics)

	assert.True(t, metrics.Alloc > 0)
	assert.True(t, metrics.RandomValue >= 0 && metrics.RandomValue <= 1)
}

func TestSendRequest(t *testing.T) {
	metrics := Metrics{
		Alloc:         1.1,
		BuckHashSys:   2.2,
		Frees:         3.3,
		GCCPUFraction: 4.4,
		GCSys:         5.5,
		HeapAlloc:     6.6,
		HeapIdle:      7.7,
		HeapInuse:     8.8,
		HeapObjects:   9.9,
		HeapReleased:  10.235,
		HeapSys:       42.31123,
		LastGC:        543.223,
		Lookups:       34113,
		MCacheInuse:   123421.5,
		MCacheSys:     563456.1324123415,
		MSpanInuse:    6.1,
		MSpanSys:      1.0,
		Mallocs:       3.5,
		NextGC:        -1234.33,
		NumForcedGC:   0.512345,
		NumGC:         -0.5611,
		OtherSys:      1.1,
		PauseTotalNs:  1,
		StackInuse:    5,
		StackSys:      0,
		Sys:           9999999,
		TotalAlloc:    23.5667,
		PollCount:     2345,
		RandomValue:   331.3356,
	}

	r := chi.NewRouter()
	r.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		metricServerTestHandler(r, t, &metrics)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	err := Report(&metrics, ts.URL)
	assert.NoError(t, err)
}

func metricServerTestHandler(r *http.Request, t *testing.T, metrics *Metrics) {
	var mr models.Metrics
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	require.NoError(t, err)
	err = json.Unmarshal(buf.Bytes(), &mr)
	require.NoError(t, err)

	switch mr.MType {
	case models.Gauge:
		v := *mr.Value

		switch mr.ID {
		case "Alloc":
			assert.Equal(t, metrics.Alloc, v)
		case "BuckHashSys":
			assert.Equal(t, metrics.BuckHashSys, v)
		case "Frees":
			assert.Equal(t, metrics.Frees, v)
		case "GCCPUFraction":
			assert.Equal(t, metrics.GCCPUFraction, v)
		case "GCSys":
			assert.Equal(t, metrics.GCSys, v)
		case "HeapAlloc":
			assert.Equal(t, metrics.HeapAlloc, v)
		case "HeapIdle":
			assert.Equal(t, metrics.HeapIdle, v)
		case "HeapInuse":
			assert.Equal(t, metrics.HeapInuse, v)
		case "HeapObjects":
			assert.Equal(t, metrics.HeapObjects, v)
		case "HeapReleased":
			assert.Equal(t, metrics.HeapReleased, v)
		case "HeapSys":
			assert.Equal(t, metrics.HeapSys, v)
		case "LastGC":
			assert.Equal(t, metrics.LastGC, v)
		case "Lookups":
			assert.Equal(t, metrics.Lookups, v)
		case "MCacheInuse":
			assert.Equal(t, metrics.MCacheInuse, v)
		case "MCacheSys":
			assert.Equal(t, metrics.MCacheSys, v)
		case "MSpanInuse":
			assert.Equal(t, metrics.MSpanInuse, v)
		case "MSpanSys":
			assert.Equal(t, metrics.MSpanSys, v)
		case "Mallocs":
			assert.Equal(t, metrics.Mallocs, v)
		case "NextGC":
			assert.Equal(t, metrics.NextGC, v)
		case "NumForcedGC":
			assert.Equal(t, metrics.NumForcedGC, v)
		case "NumGC":
			assert.Equal(t, metrics.NumGC, v)
		case "OtherSys":
			assert.Equal(t, metrics.OtherSys, v)
		case "PauseTotalNs":
			assert.Equal(t, metrics.PauseTotalNs, v)
		case "StackInuse":
			assert.Equal(t, metrics.StackInuse, v)
		case "StackSys":
			assert.Equal(t, metrics.StackSys, v)
		case "Sys":
			assert.Equal(t, metrics.Sys, v)
		case "TotalAlloc":
			assert.Equal(t, metrics.TotalAlloc, v)
		case "RandomValue":
			assert.Equal(t, metrics.RandomValue, v)
		}

	case models.Counter:
		v := *mr.Delta
		switch mr.ID {
		case "PollCount":
			assert.Equal(t, metrics.PollCount, v)
		}
	default:
		assert.Error(t, fmt.Errorf("invalid metric type: %s", mr.MType))
	}
}
