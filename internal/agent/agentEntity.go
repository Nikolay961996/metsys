package agent

import (
	"github.com/Nikolay961996/metsys/models"
	"time"
)

type Entity struct {
	doneChan chan any
}

func InitAgent() Entity {
	a := Entity{
		doneChan: make(chan any),
	}

	return a
}

func (a *Entity) Run(config *Config) {
	jobsChan := make(chan workerJob, config.SendMetricsRateLimit)
	newMetricsChan := runPollWorker(config.PollInterval, a.doneChan)
	newGopsutilMetricsChan := runPollGopsutilWorker(config.PollInterval, a.doneChan)
	runReportWorkers(a.doneChan, config.SendMetricsRateLimit, jobsChan, config.SendToServerAddress, config.KeyForSigning)
	listenMetricsAndFadeOut(a.doneChan, config.ReportInterval, newMetricsChan, newGopsutilMetricsChan, jobsChan)
}

func (a *Entity) Stop() {
	close(a.doneChan)
}

func listenMetricsAndFadeOut(doneChan chan any, period time.Duration, metricsCn <-chan Metrics, gopsutilMetricsCn <-chan MetricsGopsutil, jobsChan chan<- workerJob) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	var metrics Metrics
	var gopsutilMetrics MetricsGopsutil

	for {
		select {
		case newMetrics := <-metricsCn:
			metrics = newMetrics
		case newGMetrics := <-gopsutilMetricsCn:
			gopsutilMetrics = newGMetrics
		case <-ticker.C:
			metricsArray := createMetricsArray(&metrics)
			for _, m := range metricsArray {
				jobsChan <- workerJob{oneMetrics: m}
			}
			metrics.PollCount = 0

			gMetricsArray := createGopsutilMetricsArray(&gopsutilMetrics)
			for _, m := range gMetricsArray {
				jobsChan <- workerJob{oneMetrics: m}
			}
		case <-doneChan:
			models.Log.Warn("Listen fadeOut closed")
			return
		}
	}
}
