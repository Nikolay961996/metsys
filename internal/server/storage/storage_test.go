package storage

import (
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStorage(t *testing.T) {
	tests := []struct {
		name          string
		metricType    string
		metricName    string
		metricValue   float64
		expectedValue float64
	}{
		{"test #1", models.Gauge, "abc", 1, 1},
		{"test #2", models.Gauge, "abc", -999.3, -999.3},
		{"test #3", models.Gauge, "cdf", 12345.789, 12345.789},
		{"test #4", models.Gauge, "", 999, 999},

		{"test #5", models.Counter, "abc", 1, 1},
		{"test #6", models.Counter, "abc", -999, -998},
		{"test #7", models.Counter, "cdf", 0, 0},
		{"test #8", models.Counter, "", 999, 999},
	}

	s := NewMemStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.metricType {
			case models.Gauge:
				s.SetGauge(tt.metricName, tt.metricValue)
				v, err := s.GetGauge(tt.metricName)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValue, v)
			case models.Counter:
				s.AddCounter(tt.metricName, int64(tt.metricValue))
				v, err := s.GetCounter(tt.metricName)
				require.NoError(t, err)
				assert.Equal(t, int64(tt.expectedValue), v)
			default:
				require.Error(t, fmt.Errorf("invalid metric type: %s", tt.metricType))
			}
		})
	}
}
