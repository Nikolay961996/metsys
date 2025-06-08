package main

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/models"
	"time"
)

func main() {
	flags()
	fmt.Println("Send to", models.ServerAddress)
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

func flags() {
	flag.StringVar(&models.ServerAddress, "a", models.ServerAddress, "Metsys server address ip:port")
	r := flag.Int("r", 10, "ReportInterval in seconds")
	p := flag.Int("p", 2, "PollInterval in seconds")
	flag.Parse()

	models.ReportInterval = time.Duration(*r) * time.Second
	models.PollInterval = time.Duration(*p) * time.Second
}
