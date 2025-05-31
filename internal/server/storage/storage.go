package storage

import "errors"

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.GaugeMetrics[metricName] = value
}

func (m *MemStorage) GetGauge(metricName string) (float64, error) {
	value, ok := m.GaugeMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}

func (m *MemStorage) AddCounter(metricName string, value int64) {
	m.CounterMetrics[metricName] += value
}

func (m *MemStorage) GetCounter(metricName string) (int64, error) {
	value, ok := m.CounterMetrics[metricName]
	if !ok {
		return 0, errors.New("not Found")
	}
	return value, nil
}
