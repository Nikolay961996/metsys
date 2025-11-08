package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Nikolay961996/metsys/proto"
	"google.golang.org/grpc/metadata"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/Nikolay961996/metsys/internal/crypto"
	"github.com/Nikolay961996/metsys/utils"
	"github.com/go-resty/resty/v2"

	"github.com/Nikolay961996/metsys/models"
)

// HTTPStatusError manual error type
type HTTPStatusError struct {
	StatusCode int
}

// Error implementation
func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP error: status %d", e.StatusCode)
}

// Report to server
func Report(metrics models.Metrics, serverAddress string, keyForSigning string, publicKey *rsa.PublicKey, realIP string, GRPCClient *proto.MetricsServiceClient) error {
	if GRPCClient != nil {
		md := metadata.New(map[string]string{
			"X-Real-IP": realIP,
		})

		_, err := (*GRPCClient).UpdateMetric(metadata.NewOutgoingContext(context.Background(), md), &proto.MetricUpdateRequest{
			Id:    metrics.ID,
			Type:  metrics.MType,
			Value: *metrics.Value,
			Delta: *metrics.Delta,
		})

		if err != nil {
			models.Log.Error(fmt.Sprintf("error grpc: %s", err.Error()))
		}
	}

	client := resty.New()
	url := fmt.Sprintf("%s/update/", serverAddress)
	return sendToServer(client, url, &metrics, keyForSigning, publicKey, realIP)
}

func createMetricsArray(metrics *Metrics) []models.Metrics {
	gauges := createGaugeMetrics(metrics)
	counters := createCounterMetrics(metrics)

	return append(gauges, counters...)
}

func createGopsutilMetricsArray(metrics *MetricsGopsutil) []models.Metrics {
	gauge := map[string]float64{
		"TotalMemory":     metrics.TotalMemory,
		"FreeMemory":      metrics.FreeMemory,
		"CPUutilization1": metrics.CPUutilization1,
	}
	var arr []models.Metrics
	for k, v := range gauge {
		mr := createMetrics(models.Gauge, k, v)
		arr = append(arr, mr)
	}
	return arr
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
	switch metricType {
	case models.Gauge:
		v := metricValue.(float64)
		mr.Value = &v
	case models.Counter:
		v := metricValue.(int64)
		mr.Delta = &v
	}

	return mr
}

func sendToServer(client *resty.Client, serverURL string, metrics *models.Metrics, keyForSigning string, publicKey *rsa.PublicKey, realIP string) error {
	models.Log.Info("Sending metrics to " + serverURL)
	models.Log.Info("data: " + fmt.Sprintf("%v", metrics))

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error marshaling metrics: %s", err.Error())
	}

	sign := createSign(jsonData, keyForSigning)
	var result []byte
	if publicKey != nil {
		encryptedData, e := crypto.EncryptMessageWithPublicKey(jsonData, publicKey)
		if e != nil {
			return fmt.Errorf("error encrypting metrics: %s", e.Error())
		}
		result = encryptedData
	} else {
		// no encryption: send raw json
		result = jsonData
	}

	compressedBody, err := compressToGzip(result)
	if err != nil {
		return fmt.Errorf("error compressing metrics: %s", err.Error())
	}

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("X-Real-IP", realIP).
		SetBody(compressedBody)

	if len(sign) > 0 {
		request.SetHeader("HashSHA256", hex.EncodeToString(sign))
	}

	var resp *resty.Response
	err = utils.RetryerCon(
		func() error {
			r, e := request.Post(serverURL)
			if e == nil {
				if r.StatusCode() != http.StatusOK {
					return &HTTPStatusError{StatusCode: r.StatusCode()}
				}
				resp = r
			}
			return e
		}, func(err error) bool {
			models.Log.Warn(fmt.Sprintf("Retry error: %s", err.Error()))
			var netErr net.Error
			var netStatusErr *HTTPStatusError
			return errors.As(err, &netErr) || errors.As(err, &netStatusErr) || errors.Is(err, io.EOF)
		})
	if err != nil {
		return fmt.Errorf("failed to send metrics. %s", err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		return &HTTPStatusError{resp.StatusCode()}
	}

	return nil
}

func compressToGzip(metrics []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	var gzipWriterPool = sync.Pool{
		New: func() any {
			gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestCompression)
			return gz
		},
	}
	cw := gzipWriterPool.Get().(*gzip.Writer)
	cw.Reset(buf)
	defer func() {
		cw.Close()
		gzipWriterPool.Put(cw)
	}()
	if _, err := cw.Write(metrics); err != nil {
		return nil, fmt.Errorf("error json write: %s", err.Error())
	}
	if err := cw.Close(); err != nil {
		return nil, fmt.Errorf("error closing gzip writer: %s", err.Error())
	}
	return buf.Bytes(), nil
}

func createSign(jsonData []byte, keyForSigning string) []byte {
	if keyForSigning != "" {
		h := hmac.New(sha256.New, []byte(keyForSigning))
		h.Write(jsonData)
		return h.Sum(nil)
	}
	return nil
}
