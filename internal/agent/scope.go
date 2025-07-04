package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/go-resty/resty/v2"
	"net/http"
)

func Report(metrics *Metrics, serverAddress string) error {
	client := resty.New().
		SetTimeout(models.SendMetricTimeout)

	err := sendGaugeMetrics(client, serverAddress, metrics)
	if err != nil {
		return err
	}

	err = sendCounterMetrics(client, serverAddress, metrics)
	if err != nil {
		return err
	}

	return nil
}

func sendGaugeMetrics(client *resty.Client, serverAddress string, metrics *Metrics) error {
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

	for k, v := range gauge {
		err := sendMetricJSON(client, serverAddress, models.Gauge, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendCounterMetrics(client *resty.Client, serverAddress string, metrics *Metrics) error {
	counter := map[string]int64{
		"PollCount": metrics.PollCount,
	}

	for k, v := range counter {
		err := sendMetricJSON(client, serverAddress, models.Counter, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendMetricJSON(client *resty.Client, serverAddress string, metricType string, metricName string, metricValue any) error {
	mr := createMetrics(metricType, metricName, metricValue)
	err := sendToServer(client, serverAddress, mr)
	return err
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

func sendToServer(client *resty.Client, serverAddress string, mr models.Metrics) error {
	body, err := compressToGzip(mr)
	if err != nil {
		return fmt.Errorf("error compressing metrics: %s", err.Error())
	}
	url := fmt.Sprintf("%s/update/", serverAddress)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(body).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric (%s). %s", mr.ID, err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed status to send metrics: %d", resp.StatusCode())
	}

	return nil
}

func compressToGzip(metrics models.Metrics) ([]byte, error) {
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
