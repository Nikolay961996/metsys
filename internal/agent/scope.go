package agent

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"io"
	"net/http"
)

func Report(metrics *Metrics) error {
	client := &http.Client{
		Timeout: models.SendMetricTimeout,
	}

	err := sendGaugeMetrics(client, metrics)
	if err != nil {
		return err
	}

	err = sendCounterMetrics(client, metrics)
	if err != nil {
		return err
	}

	return nil
}

func sendGaugeMetrics(client *http.Client, metrics *Metrics) error {
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
		err := sendMetric(client, models.Gauge, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendCounterMetrics(client *http.Client, metrics *Metrics) error {
	counter := map[string]int64{
		"PollCount": metrics.PollCount,
	}

	for k, v := range counter {
		err := sendMetric(client, models.Counter, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendMetric(client *http.Client, metricType string, metricName string, metricValue any) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", models.ServerAddress, metricType, metricName, metricValue)
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("failed to send metric (%s) = %v", metricName, metricValue)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed status to send metrics: %d", resp.StatusCode)
	}
	return nil
}
