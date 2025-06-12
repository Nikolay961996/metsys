package main

import (
	"flag"
	"fmt"
	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/Nikolay961996/metsys/models"
	"github.com/caarlos0/env/v6"
	"os"
	"strings"
	"time"
)

func main() {
	flags()
	envs()

	fmt.Println("Send to", models.SendToServerAddress)
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
			err := agent.Report(metrics, models.SendToServerAddress)
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
	flag.StringVar(&models.SendToServerAddress, "a", "http://localhost:8080", "Metsys server address ip:port")
	r := flag.Int("r", 10, "ReportInterval in seconds")
	p := flag.Int("p", 2, "PollInterval in seconds")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Printf("Unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}

	models.SendToServerAddress = fixProtocolPrefixAddress(models.SendToServerAddress)
	models.ReportInterval = time.Duration(*r) * time.Second
	models.PollInterval = time.Duration(*p) * time.Second
}

func envs() {
	var c struct {
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
	}
	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	if c.Address != "" {
		models.SendToServerAddress = fixProtocolPrefixAddress(c.Address)
	}
	if c.ReportInterval != 0 {
		models.ReportInterval = time.Duration(c.ReportInterval) * time.Second
	}
	if c.PollInterval != 0 {
		models.PollInterval = time.Duration(c.PollInterval) * time.Second
	}
}

func fixProtocolPrefixAddress(addr string) string {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	addr = strings.TrimRight(addr, "/")

	return addr
}
