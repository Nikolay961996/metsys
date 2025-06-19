package agent

import (
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

func sendMetric(client *resty.Client, serverAddress string, metricType string, metricName string, metricValue any) error {
	url := fmt.Sprintf("%s/update/{metricType}/{metricName}/{metricValue}", serverAddress)
	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"metricType":  metricType,
			"metricName":  metricName,
			"metricValue": fmt.Sprintf("%v", metricValue),
		}).Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric (%s) = %v. %s", metricName, metricValue, err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed status to send metrics: %d", resp.StatusCode())
	}
	return nil
}

func sendMetricJSON(client *resty.Client, serverAddress string, metricType string, metricName string, metricValue any) error {
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

	url := fmt.Sprintf("%s/update/", serverAddress)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(mr).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric (%s) = %v. %s", metricName, metricValue, err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed status to send metrics: %d", resp.StatusCode())
	}
	return nil
}
