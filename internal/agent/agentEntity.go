package agent

import (
	"fmt"
	"time"
)

type Entity struct {
	metrics      Metrics
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
	config       *Config
}

func InitAgent(c *Config) Entity {
	a := Entity{
		config:       c,
		pollTicker:   time.NewTicker(c.PollInterval),
		reportTicker: time.NewTicker(c.ReportInterval),
	}
	return a
}

func (a *Entity) Run() {
	for {
		select {
		case <-a.pollTicker.C:
			Poll(&a.metrics)
			fmt.Println("Metrics poll")
		case <-a.reportTicker.C:
			err := Report(&a.metrics, a.config.SendToServerAddress, a.config.KeyForSigning)
			if err != nil {
				fmt.Println("Error while reporting: ", err)
			} else {
				a.metrics.PollCount = 0
				fmt.Println("Metrics reported")
			}
		}
	}
}

func (a *Entity) Stop() {
	a.pollTicker.Stop()
	a.reportTicker.Stop()
}
