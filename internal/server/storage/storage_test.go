package storage

import (
	"encoding/json"
	"fmt"
	"github.com/Nikolay961996/metsys/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
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

	s := NewFileStorage("tst", false, false)

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

func TestSyncSaveFile(t *testing.T) {
	file := "tst.db"
	s := NewFileStorage(file, true, false)
	s.SetGauge("aaa", 123.4)
	s.AddCounter("bbb", 987)

	bytes, err := os.ReadFile(file)
	require.NoError(t, err)
	var s2 MemStorage
	err = json.Unmarshal(bytes, &s2)
	require.NoError(t, err)
	g1, err := s.GetGauge("aaa")
	require.NoError(t, err)
	c1, err := s.GetCounter("bbb")
	require.NoError(t, err)
	g2, err := s2.GetGauge("aaa")
	require.NoError(t, err)
	c2, err := s2.GetCounter("bbb")
	require.NoError(t, err)
	assert.Equal(t, g1, g2)
	assert.Equal(t, c1, c2)
}

func TestSaveFile(t *testing.T) {
	file := "tst.db"
	s := NewFileStorage(file, false, false)
	s.SetGauge("aaa", 123.4)
	s.AddCounter("bbb", 987)
	s.TryFlushToFile()

	bytes, err := os.ReadFile(file)
	require.NoError(t, err)
	var s2 MemStorage
	err = json.Unmarshal(bytes, &s2)
	require.NoError(t, err)
	g1, err := s.GetGauge("aaa")
	require.NoError(t, err)
	g2, err := s2.GetGauge("aaa")
	require.NoError(t, err)
	assert.Equal(t, g1, g2)
}

func TestLoadFile(t *testing.T) {
	file := "tst.db"
	s := MemStorage{
		GaugeMetrics: map[string]float64{
			"aaa1": 13.3,
		},
		CounterMetrics: map[string]int64{
			"ccc3": 888,
		},
	}

	bytes, err := json.Marshal(s)
	require.NoError(t, err)
	err = os.WriteFile(file, bytes, 0666)
	require.NoError(t, err)

	s2 := NewFileStorage(file, false, true)
	g1, err := s.GetGauge("aaa1")
	require.NoError(t, err)
	g2, err := s2.GetGauge("aaa1")
	require.NoError(t, err)
	assert.Equal(t, g1, g2)

	c1, err := s.GetCounter("ccc3")
	require.NoError(t, err)
	c2, err := s2.GetCounter("ccc3")
	require.NoError(t, err)
	assert.Equal(t, c1, c2)
}
