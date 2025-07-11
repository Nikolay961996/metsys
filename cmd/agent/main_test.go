package main

import (
	"github.com/Nikolay961996/metsys/internal/agent"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimers(t *testing.T) {
	c := agent.DefaultConfig()
	pollTicker := time.NewTicker(c.PollInterval)
	reportTicker := time.NewTicker(c.ReportInterval)
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

	time.Sleep(c.ReportInterval + time.Second)
	pollTicker.Stop()
	reportTicker.Stop()
}
