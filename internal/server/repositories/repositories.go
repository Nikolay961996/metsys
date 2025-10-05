// Package repositories consist storage interface
package repositories

import "context"

type MetricDto struct {
	Name  string
	Type  string
	Value string
}

type Storage interface {
	SetGauge(metricName string, value float64)
	GetGauge(metricName string) (float64, error)
	AddCounter(metricName string, value int64)
	GetCounter(metricName string) (int64, error)
	GetAll() []MetricDto
	Close()
	PingContext(ctx context.Context) error
	StartTransaction(ctx context.Context) error
	CommitTransaction() error
}
