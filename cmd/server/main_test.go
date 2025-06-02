package main

import (
	"fmt"
	"github.com/Nikolay961996/metsys/internal/server/handlers"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestPositiveServer(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{"test #1", http.MethodPost, "/update/gauge/memory/12.34", want{http.StatusOK}},
		{"test #2", http.MethodPost, "/update/gauge/memory/0", want{http.StatusOK}},
		{"test #3", http.MethodPost, "/update/gauge/CP/-9999.999", want{http.StatusOK}},
		{"test #4", http.MethodPost, "/update/counter/processes/12", want{http.StatusOK}},
		{"test #5", http.MethodPost, "/update/counter/memory/-99", want{http.StatusOK}},
		{"test #6", http.MethodPost, "/update/counter/memory/0", want{http.StatusOK}},
	}
	s := storage.NewMemStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)
			request.Header.Set("Content-Type", "text/plain")
			response := httptest.NewRecorder()

			h := handlers.UpdateMetricHandler(s)
			h(response, request)
			result := response.Result()
			defer result.Body.Close()

			params := strings.Split(tt.url, "/")
			metricType := params[2]
			metricName := params[3]
			metricValue := params[4]
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			switch metricType {
			case models.Gauge:
				v, err := s.GetGauge(metricName)
				require.NoError(t, err)
				assert.Equal(t, metricValue, strconv.FormatFloat(v, 'f', -1, 64))
			case models.Counter:
				_, err := s.GetCounter(metricName)
				require.NoError(t, err)
			default:
				require.Error(t, fmt.Errorf("unknown metric type: %s", metricType))
			}
		})
	}
}

func TestNegativeServer(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		method      string
		url         string
		contentType string
		want        want
	}{
		{"test #1", http.MethodGet, "/update/gauge/test/12.34", "text/plain", want{http.StatusMethodNotAllowed}},
		{"test #2", http.MethodPost, "/update/gauge/test/12.34", "application/json", want{http.StatusUnsupportedMediaType}},
		{"test #3", http.MethodPost, "/update/some/test/12.34", "text/plain", want{http.StatusBadRequest}},
		{"test #4", http.MethodPost, "/update/gauge/memory", "text/plain", want{http.StatusNotFound}},
		{"test #5", http.MethodPost, "/update/gauge/memory/", "text/plain", want{http.StatusBadRequest}},
		{"test #6", http.MethodPost, "/update//memory/1", "text/plain", want{http.StatusBadRequest}},
		{"test #7", http.MethodPost, "/update/gauge//1", "text/plain", want{http.StatusNotFound}},
		{"test #8", http.MethodPost, "/update///", "text/plain", want{http.StatusNotFound}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)
			request.Header.Set("Content-Type", tt.contentType)
			response := httptest.NewRecorder()

			h := handlers.UpdateMetricHandler(nil)
			h(response, request)
			result := response.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
