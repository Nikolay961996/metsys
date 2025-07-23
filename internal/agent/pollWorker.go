package agent

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"time"
)

func runPollWorker(period time.Duration, doneChan chan any) chan Metrics {
	ticker := time.NewTicker(period)
	outCh := make(chan Metrics)

	go func() {
		m := Metrics{}
		defer ticker.Stop()
		defer close(outCh)
		for {
			select {
			case <-ticker.C:
				Poll(&m)
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
