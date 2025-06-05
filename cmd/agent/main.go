package main

import (
	"fmt"
	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/models"
	"time"
)

func main() {
	pollTicker := time.NewTicker(models.PollInterval)
	reportTicker := time.NewTicker(models.ReportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	var metrics agent.Metrics
	run(pollTicker, reportTicker, &metrics)
}

func run(pollTicker *time.Ticker, reportTicker *time.Ticker, metrics *agent.Metrics) {
	for {
		select {
		case <-pollTicker.C:
			agent.Poll(metrics)
			fmt.Println("Metrics poll")
		case <-reportTicker.C:
			err := agent.Report(metrics, models.ServerAddress)
			if err != nil {
				fmt.Println("Error while reporting: ", err)
			} else {
				metrics.PollCount = 0
				fmt.Println("Metrics reported")
			}
		}
	}
}
