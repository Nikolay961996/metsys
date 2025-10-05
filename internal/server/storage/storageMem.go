// Package storage in memory storage
package storage

import (
	"context"
	"errors"
	"strconv"

	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/models"
)

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewMemStorage() *MemStorage {
	s := MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	return &s
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

func (m *MemStorage) GetAll() []repositories.MetricDto {
	var r []repositories.MetricDto
	for k, v := range m.GaugeMetrics {
		r = append(r, repositories.MetricDto{
			Name:  k,
			Type:  models.Gauge,
			Value: strconv.FormatFloat(v, 'f', -1, 64),
		})
	}
	for k, v := range m.CounterMetrics {
		r = append(r, repositories.MetricDto{
			Name:  k,
			Type:  models.Counter,
			Value: strconv.FormatInt(v, 10),
		})
	}
	return r
}

func (m *MemStorage) Close() {}

func (m *MemStorage) PingContext(_ context.Context) error {
	return nil
}

func (m *MemStorage) StartTransaction(_ context.Context) error {
	return nil
}
func (m *MemStorage) CommitTransaction() error {
	return nil
}
