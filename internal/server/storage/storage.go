package storage

import (
	"errors"
	"github.com/Nikolay961996/metsys/models"
	"strconv"
)

type MetricDto struct {
	Name  string
	Type  string
	Value string
}

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

func (m *MemStorage) GetAll() []MetricDto {
	var r []MetricDto
	for k, v := range m.GaugeMetrics {
		r = append(r, MetricDto{
			Name:  k,
			Type:  models.Gauge,
			Value: strconv.FormatFloat(v, 'f', -1, 64),
		})
	}
	for k, v := range m.CounterMetrics {
		r = append(r, MetricDto{
			Name:  k,
			Type:  models.Counter,
			Value: strconv.FormatInt(v, 10),
		})
	}
	return r
}
