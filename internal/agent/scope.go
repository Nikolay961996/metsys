package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-resty/resty/v2"
	"io"
	"net"
	"net/http"
)

type HTTPStatusError struct {
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP error: status %d", e.StatusCode)
}

func Report(metrics *Metrics, serverAddress string) error {
	client := resty.New().
		SetTimeout(models.SendMetricTimeout)

	allMetrics := createMetricsArray(metrics)
	if len(allMetrics) == 0 {
		return nil
	}
	url := fmt.Sprintf("%s/updates/", serverAddress)

	err := models.RetryerCon(
		func() error {
			return sendToServer(client, url, allMetrics)
		}, func(err error) bool {
			var netErr net.Error
			var netStatusErr *HTTPStatusError
			return errors.As(err, &netErr) || errors.As(err, &netStatusErr) || errors.Is(err, io.EOF)
		})
	if err != nil {
		return err
	}

	return nil
}

func createMetricsArray(metrics *Metrics) []models.Metrics {
	gauges := createGaugeMetrics(metrics)
	counters := createCounterMetrics(metrics)

	return append(gauges, counters...)
}

func createGaugeMetrics(metrics *Metrics) []models.Metrics {
	gauge := map[string]float64{
		"Alloc":         metrics.Alloc,
		"BuckHashSys":   metrics.BuckHashSys,
		"Frees":         metrics.Frees,
		"GCCPUFraction": metrics.GCCPUFraction,
		"GCSys":         metrics.GCSys,
		"HeapAlloc":     metrics.HeapAlloc,
		"HeapIdle":      metrics.HeapIdle,
		"HeapInuse":     metrics.HeapInuse,
		"HeapObjects":   metrics.HeapObjects,
		"HeapReleased":  metrics.HeapReleased,
		"HeapSys":       metrics.HeapSys,
		"LastGC":        metrics.LastGC,
		"Lookups":       metrics.Lookups,
		"MCacheInuse":   metrics.MCacheInuse,
		"MCacheSys":     metrics.MCacheSys,
		"MSpanInuse":    metrics.MSpanInuse,
		"MSpanSys":      metrics.MSpanSys,
		"Mallocs":       metrics.Mallocs,
		"NextGC":        metrics.NextGC,
		"NumForcedGC":   metrics.NumForcedGC,
		"NumGC":         metrics.NumGC,
		"OtherSys":      metrics.OtherSys,
		"PauseTotalNs":  metrics.PauseTotalNs,
		"StackInuse":    metrics.StackInuse,
		"StackSys":      metrics.StackSys,
		"Sys":           metrics.Sys,
		"TotalAlloc":    metrics.TotalAlloc,
		"RandomValue":   metrics.RandomValue,
	}
	var arr []models.Metrics
	for k, v := range gauge {
		mr := createMetrics(models.Gauge, k, v)
		arr = append(arr, mr)
	}
	return arr
}

func createCounterMetrics(metrics *Metrics) []models.Metrics {
	counter := map[string]int64{
		"PollCount": metrics.PollCount,
	}

	var arr []models.Metrics
	for k, v := range counter {
		mr := createMetrics(models.Counter, k, v)
		arr = append(arr, mr)
	}

	return arr
}

func createMetrics(metricType string, metricName string, metricValue any) models.Metrics {
	mr := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}
	if metricType == models.Gauge {
		v := metricValue.(float64)
		mr.Value = &v
	} else if metricType == models.Counter {
		v := metricValue.(int64)
		mr.Delta = &v
	}

	return mr
}

func sendToServer(client *resty.Client, serverURL string, metrics any) error {
	body, err := compressToGzip(metrics)
	if err != nil {
		return fmt.Errorf("error compressing metrics: %s", err.Error())
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		Post(serverURL)

	if err != nil {
		return fmt.Errorf("failed to send metrics. %s", err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return &HTTPStatusError{resp.StatusCode()}
	}

	return nil
}

func compressToGzip(metrics any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	cw := gzip.NewWriter(buf)
	d, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("error json marshaling: %s", err.Error())
	}

	if _, err := cw.Write(d); err != nil {
		return nil, fmt.Errorf("error json write: %s", err.Error())
	}
	if err := cw.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %s", err.Error())
	}

	return buf.Bytes(), nil
}
