package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/internal/server/storage"
	"github.com/Nikolay961996/metsys/models"
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
	s := storage.NewFileStorage("tst", 5*time.Second, false)

	ts := httptest.NewServer(router.MetricsRouterWithServer(s, ""))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "text/plain")

			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			params := strings.Split(tt.url, "/")
			metricType := params[2]
			metricName := params[3]
			metricValue := params[4]
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
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
		//{"test #2", http.MethodPost, "/update/gauge/test/12.34", "application/json", want{http.StatusUnsupportedMediaType}},
		{"test #3", http.MethodPost, "/update/some/test/12.34", "text/plain", want{http.StatusBadRequest}},
		{"test #4", http.MethodPost, "/update/gauge/memory", "text/plain", want{http.StatusNotFound}},
		{"test #5", http.MethodPost, "/update/gauge/memory/", "text/plain", want{http.StatusNotFound}},
		{"test #6", http.MethodPost, "/update//memory/1", "text/plain", want{http.StatusBadRequest}},
		{"test #7", http.MethodPost, "/update/gauge//1", "text/plain", want{http.StatusNotFound}},
		{"test #8", http.MethodPost, "/update///", "text/plain", want{http.StatusNotFound}},
	}

	ts := httptest.NewServer(router.MetricsRouterTest())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)
			request.Header.Set("Content-Type", tt.contentType)

			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestServer(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		method string
		url    string
		name   string
		want   want
	}{
		{http.MethodPost, "/update/gauge/memory/12.34", "test #1", want{http.StatusOK}},
		{http.MethodPost, "/update/counter/health/-99", "test #2", want{http.StatusOK}},
		{http.MethodGet, "/update/gauge/test/12.34", "test #3", want{http.StatusMethodNotAllowed}},
		{http.MethodPost, "/update/undefinedType/test/12.34", "test #4", want{http.StatusBadRequest}},
		{http.MethodPost, "/update/gauge/memory", "test #5", want{http.StatusNotFound}},
		{http.MethodPost, "/update/gauge/memory/", "test #6", want{http.StatusNotFound}},
		{http.MethodPost, "/update//memory/1", "test #7", want{http.StatusBadRequest}},
		{http.MethodPost, "/update/gauge//1", "test #8", want{http.StatusNotFound}},
		{http.MethodPost, "/update///", "test #9", want{http.StatusNotFound}},
	}

	ts := httptest.NewServer(router.MetricsRouterTest())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "text/plain")
			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestGetMetric(t *testing.T) {
	type want struct {
		value      string
		statusCode int
	}
	tests := []struct {
		method string
		url    string
		name   string
		want   want
	}{
		{http.MethodPost, "/update/gauge/memory/12.34", "test #1", want{"", http.StatusOK}},
		{http.MethodGet, "/value/gauge/memory", "test #2", want{"12.34", http.StatusOK}},
		{http.MethodGet, "/value/gauge/memory2", "test #3", want{"", http.StatusNotFound}},
		{http.MethodPost, "/update/gauge/memory/99.7654", "test #4", want{"", http.StatusOK}},
		{http.MethodGet, "/value/gauge/memory", "test #5", want{"99.7654", http.StatusOK}},
		{http.MethodGet, "/value/counter/cp", "test #6", want{"", http.StatusNotFound}},
		{http.MethodPost, "/update/counter/cp/123", "test #7", want{"", http.StatusOK}},
		{http.MethodGet, "/value/counter/cp", "test #8", want{"123", http.StatusOK}},
		{http.MethodPost, "/update/counter/cp/100", "test #9", want{"", http.StatusOK}},
		{http.MethodGet, "/value/counter/cp", "test #10", want{"223", http.StatusOK}},
	}

	ts := httptest.NewServer(router.MetricsRouterTest())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, ts.URL+tt.url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "text/plain")
			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if tt.method == http.MethodGet && resp.StatusCode == http.StatusOK {
				r, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.value, string(r))
			}
		})
	}
}

func TestJSONSupport(t *testing.T) {
	type want struct {
		value      float64
		delta      int64
		statusCode int
	}
	tests := []struct {
		method string
		url    string
		name   string
		body   models.Metrics
		value  float64
		delta  int64
		want   want
	}{
		{http.MethodPost, "/update/", "test #1", models.Metrics{ID: "memory", MType: models.Gauge}, 12.34, 0, want{12.34, 0, http.StatusOK}},
		{http.MethodPost, "/value/", "test #2", models.Metrics{ID: "memory", MType: models.Gauge}, 0, 0, want{12.34, 0, http.StatusOK}},
		{http.MethodPost, "/update/", "test #3", models.Metrics{ID: "cp", MType: models.Counter}, 0, 123, want{0, 123, http.StatusOK}},
		{http.MethodPost, "/value/", "test #4", models.Metrics{ID: "cp", MType: models.Counter}, 0, 0, want{0, 123, http.StatusOK}},
		{http.MethodPost, "/update/", "test #5", models.Metrics{ID: "cp", MType: models.Counter}, 0, 100, want{0, 223, http.StatusOK}},
		{http.MethodPost, "/value/", "test #6", models.Metrics{ID: "cp", MType: models.Counter}, 0, 0, want{0, 223, http.StatusOK}},
	}

	ts := httptest.NewServer(router.MetricsRouterTest())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.url, "update") {
				if tt.body.MType == models.Gauge {
					tt.body.Value = &tt.value
				} else {
					tt.body.Delta = &tt.delta
				}
			}

			var buf bytes.Buffer
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)
			_, err = buf.Write(body)
			require.NoError(t, err)

			request, err := http.NewRequest(tt.method, ts.URL+tt.url, &buf)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")
			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				buf.Reset()
				_, err = buf.ReadFrom(resp.Body)
				require.NoError(t, err)
				defer resp.Body.Close()
				var result models.Metrics
				err = json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)

				assert.Equal(t, tt.body.ID, result.ID)
				assert.Equal(t, tt.body.MType, result.MType)
				if tt.body.MType == models.Gauge {
					assert.Equal(t, tt.want.value, *result.Value)
				} else {
					assert.Equal(t, tt.want.delta, *result.Delta)
				}
			}
		})
	}
}
