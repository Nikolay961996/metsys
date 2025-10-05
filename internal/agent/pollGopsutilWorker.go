package agent

import (
	"context"
	"time"

	"github.com/Nikolay961996/metsys/models"
)

func runPollGopsutilWorker(period time.Duration, doneCtx context.Context) chan MetricsGopsutil {
	ticker := time.NewTicker(period)
	outCh := make(chan MetricsGopsutil, 3)

	go func() {
		m := MetricsGopsutil{}
		defer ticker.Stop()
		defer close(outCh)
		for {
			select {
			case <-ticker.C:
				PollGopsutil(&m)
				outCh <- m
				models.Log.Info("Metrics poll")
			case <-doneCtx.Done():
				models.Log.Warn("PollWorker get done signal")
				return
			}
		}
	}()

	return outCh
}
