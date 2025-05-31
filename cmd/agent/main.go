package main

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Metrics struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	PollCount     int64
	RandomValue   float64
}

func main() {
	pollTicker := time.NewTicker(models.PollInterval)
	reportTicker := time.NewTicker(models.ReportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	var metrics Metrics
	for {
		select {
		case <-pollTicker.C:
			poll(&metrics)
			fmt.Println("Metrics poll")
		case <-reportTicker.C:
			err := report(&metrics)
			if err != nil {
				fmt.Println("Error while reporting: ", err)
			} else {
				metrics.PollCount = 0
				fmt.Println("Metrics reported")
			}
		}
	}
}

func report(metrics *Metrics) error {
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
	guage := map[string]float64{
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

	for k, v := range guage {
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
		return fmt.Errorf("Failed to send metric (%s) = %v", metricName, metricValue)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed status to send metrics: %d", resp.StatusCode)
	}
	return nil
}

func poll(metrics *Metrics) {
	var stats runtime.MemStats
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	runtime.ReadMemStats(&stats)

	metrics.Alloc = float64(stats.Alloc)
	metrics.BuckHashSys = float64(stats.BuckHashSys)
	metrics.Frees = float64(stats.Frees)
	metrics.GCCPUFraction = stats.GCCPUFraction
	metrics.GCSys = float64(stats.GCSys)
	metrics.HeapAlloc = float64(stats.HeapAlloc)
	metrics.HeapIdle = float64(stats.HeapIdle)
	metrics.HeapInuse = float64(stats.HeapInuse)
	metrics.HeapObjects = float64(stats.HeapObjects)
	metrics.HeapReleased = float64(stats.HeapReleased)
	metrics.HeapSys = float64(stats.HeapSys)
	metrics.LastGC = float64(stats.LastGC)
	metrics.Lookups = float64(stats.Lookups)
	metrics.MCacheInuse = float64(stats.MCacheInuse)
	metrics.MCacheSys = float64(stats.MCacheSys)
	metrics.MSpanInuse = float64(stats.MSpanInuse)
	metrics.MSpanSys = float64(stats.MSpanSys)
	metrics.Mallocs = float64(stats.Mallocs)
	metrics.NextGC = float64(stats.NextGC)
	metrics.NumForcedGC = float64(stats.NumForcedGC)
	metrics.NumGC = float64(stats.NumGC)
	metrics.OtherSys = float64(stats.OtherSys)
	metrics.PauseTotalNs = float64(stats.PauseTotalNs)
	metrics.StackInuse = float64(stats.StackInuse)
	metrics.StackSys = float64(stats.StackSys)
	metrics.Sys = float64(stats.Sys)
	metrics.TotalAlloc = float64(stats.TotalAlloc)
	metrics.PollCount++
	metrics.RandomValue = random.Float64()
}
