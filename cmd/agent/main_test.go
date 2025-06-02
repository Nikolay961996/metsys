package main

import (
	"github.com/Nikolay961996/metsys/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimers(t *testing.T) {
	pollTicker := time.NewTicker(models.PollInterval)
	reportTicker := time.NewTicker(models.ReportInterval)
	var pollCount int64

	go func() {
		for {
			select {
			case <-pollTicker.C:
				pollCount++
			case <-reportTicker.C:
				assert.NotZero(t, pollCount)
			}
		}
	}()

	time.Sleep(models.ReportInterval + time.Second)
	pollTicker.Stop()
	reportTicker.Stop()
}
