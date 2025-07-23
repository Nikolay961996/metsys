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
	runReportWorkers(a.doneChan, config.SendMetricsRateLimit, jobsChan, config.SendToServerAddress, config.KeyForSigning)
	listenMetricsAndFadeOut(a.doneChan, config.ReportInterval, newMetricsChan, jobsChan)
}

func (a *Entity) Stop() {
	close(a.doneChan)
}

func listenMetricsAndFadeOut(doneChan chan any, period time.Duration, metricsCn <-chan Metrics, jobsChan chan<- workerJob) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	var metrics Metrics

	for {
		select {
		case newMetrics := <-metricsCn:
			metrics = newMetrics
		case <-ticker.C:
			metricsArray := createMetricsArray(&metrics)
			for _, m := range metricsArray {
				jobsChan <- workerJob{oneMetrics: m}
			}
			metrics.PollCount = 0
		case <-doneChan:
			models.Log.Warn("Listen fadeOut closed")
			return
		}
	}
}
