package agent

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"time"
)

func runPollGopsutilWorker(period time.Duration, doneChan chan any) chan MetricsGopsutil {
	ticker := time.NewTicker(period)
	outCh := make(chan MetricsGopsutil)

	go func() {
		m := MetricsGopsutil{}
		defer ticker.Stop()
		defer close(outCh)
		for {
			select {
			case <-ticker.C:
				PollGopsutil(&m)
				outCh <- m
				fmt.Println("Metrics poll")
			case <-doneChan:
				models.Log.Warn("PollWorker get done signal")
				return
			}
		}
	}()

	return outCh
}
