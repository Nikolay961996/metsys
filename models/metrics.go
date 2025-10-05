// Package models global consts and init functions
package models

import (
	"time"

	"go.uber.org/zap"
)

// Metric types
const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	Log = zap.NewNop()
)

const (
	SendMetricTimeout = time.Minute // period for send timeout
	MaxErrRetryCount  = 4           // max expenecial back-off (not included)
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
	//Hash  string   `json:"hash,omitempty"`
}

// Initialize logger
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
