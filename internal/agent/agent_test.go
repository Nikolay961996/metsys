package agent

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricsCollect(t *testing.T) {
	var metrics Metrics
	Poll(&metrics)

	assert.True(t, metrics.Alloc > 0)
	assert.True(t, metrics.RandomValue >= 0 && metrics.RandomValue <= 1)
}

func TestSendRequest(t *testing.T) {

}
