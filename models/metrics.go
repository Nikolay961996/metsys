package models

import "time"

const (
	Counter = "counter"
	Gauge   = "gauge"
)

const (
	PollInterval      = 2 * time.Second
	ReportInterval    = 10 * time.Second
	SendMetricTimeout = 5 * time.Second
	ServerAddress     = "http://localhost:8080"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
