package agent

import (
	"testing"
)

func BenchmarkAgentPollingProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := Metrics{}
		Poll(&m)
		_ = createMetricsArray(&m)
	}
}

func BenchmarkAgentGopsutilPollingProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := MetricsGopsutil{}
		PollGopsutil(&m)
		_ = createGopsutilMetricsArray(&m)
	}
}
