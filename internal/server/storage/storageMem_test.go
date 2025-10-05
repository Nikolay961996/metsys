package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Nikolay961996/metsys/internal/server/storage"
)

// TestMemStorage_GaugeOperations тестирует операции с gauge метриками
func TestMemStorage_GaugeOperations(t *testing.T) {
	s := storage.NewMemStorage()

	s.SetGauge("temperature", 23.5)
	value, err := s.GetGauge("temperature")
	if err != nil {
		t.Errorf("GetGauge failed: %v", err)
	}
	if value != 23.5 {
		t.Errorf("Expected 23.5, got %f", value)
	}

	s.SetGauge("temperature", 25.0)
	value, err = s.GetGauge("temperature")
	if err != nil {
		t.Errorf("GetGauge after overwrite failed: %v", err)
	}
	if value != 25.0 {
		t.Errorf("Expected 25.0 after overwrite, got %f", value)
	}

	_, err = s.GetGauge("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent metric")
	}
}

// TestMemStorage_CounterOperations тестирует операции с counter метриками
func TestMemStorage_CounterOperations(t *testing.T) {
	s := storage.NewMemStorage()

	s.AddCounter("requests", 10)
	value, err := s.GetCounter("requests")
	if err != nil {
		t.Errorf("GetCounter failed: %v", err)
	}
	if value != 10 {
		t.Errorf("Expected 10, got %d", value)
	}

	s.AddCounter("requests", 5)
	value, err = s.GetCounter("requests")
	if err != nil {
		t.Errorf("GetCounter after increment failed: %v", err)
	}
	if value != 15 {
		t.Errorf("Expected 15 after increment, got %d", value)
	}

	_, err = s.GetCounter("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent metric")
	}
}

// TestMemStorage_GetAll тестирует получение всех метрик
func TestMemStorage_GetAll(t *testing.T) {
	s := storage.NewMemStorage()

	s.SetGauge("cpu_usage", 75.5)
	s.SetGauge("memory_usage", 45.2)
	s.AddCounter("requests", 100)
	s.AddCounter("errors", 5)

	metrics := s.GetAll()
	if len(metrics) != 4 {
		t.Errorf("Expected 4 metrics, got %d", len(metrics))
	}

	var gaugeCount, counterCount int
	for _, metric := range metrics {
		switch metric.Type {
		case "gauge":
			gaugeCount++
		case "counter":
			counterCount++
		}
	}

	if gaugeCount != 2 {
		t.Errorf("Expected 2 gauge metrics, got %d", gaugeCount)
	}
	if counterCount != 2 {
		t.Errorf("Expected 2 counter metrics, got %d", counterCount)
	}
}

// TestMemStorage_Ping тестирует проверку соединения
func TestMemStorage_Ping(t *testing.T) {
	s := storage.NewMemStorage()
	ctx := context.Background()

	err := s.PingContext(ctx)
	if err != nil {
		t.Errorf("PingContext failed: %v", err)
	}
}

// TestMemStorage_ConcurrentAccess тестирует конкурентный доступ
func TestMemStorage_ConcurrentAccess(t *testing.T) {
	s := storage.NewMemStorage()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			s.SetGauge("concurrent_metric", float64(index))
			s.AddCounter("concurrent_counter", int64(index))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	_, err := s.GetGauge("concurrent_metric")
	if err != nil {
		t.Errorf("GetGauge after concurrent access failed: %v", err)
	}

	_, err = s.GetCounter("concurrent_counter")
	if err != nil {
		t.Errorf("GetCounter after concurrent access failed: %v", err)
	}
}

// ExampleMemStorage демонстрирует базовое использование MemStorage
func ExampleMemStorage() {
	storage := storage.NewMemStorage()

	storage.SetGauge("temperature", 23.5)
	value, _ := storage.GetGauge("temperature")
	fmt.Printf("Temperature: %.1f\n", value)

	storage.AddCounter("requests", 10)
	storage.AddCounter("requests", 5)
	count, _ := storage.GetCounter("requests")
	fmt.Printf("Requests: %d\n", count)

	// Output:
	// Temperature: 23.5
	// Requests: 15
}

// ExampleMemStorage_GetAll демонстрирует получение всех метрик
func ExampleMemStorage_GetAll() {
	storage := storage.NewMemStorage()

	storage.SetGauge("cpu", 50.0)
	storage.SetGauge("memory", 75.5)
	storage.AddCounter("requests", 100)

	metrics := storage.GetAll()
	fmt.Printf("Total metrics: %d\n", len(metrics))

	// Output:
	// Total metrics: 3
}

// ExampleMemStorage_errorHandling демонстрирует обработку ошибок
func ExampleMemStorage_errorHandling() {
	storage := storage.NewMemStorage()

	_, err := storage.GetGauge("nonexistent")
	if err != nil {
		fmt.Println("Metric not found")
	}

	_, err = storage.GetCounter("nonexistent")
	if err != nil {
		fmt.Println("Counter not found")
	}

	// Output:
	// Metric not found
	// Counter not found
}

// BenchmarkMemStorage_SetGauge бенчмарк для SetGauge
func BenchmarkMemStorage_SetGauge(b *testing.B) {
	s := storage.NewMemStorage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.SetGauge("benchmark_metric", float64(i))
	}
}

// BenchmarkMemStorage_GetGauge бенчмарк для GetGauge
func BenchmarkMemStorage_GetGauge(b *testing.B) {
	s := storage.NewMemStorage()
	s.SetGauge("benchmark_metric", 123.45)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GetGauge("benchmark_metric")
	}
}

// BenchmarkMemStorage_AddCounter бенчмарк для AddCounter
func BenchmarkMemStorage_AddCounter(b *testing.B) {
	s := storage.NewMemStorage()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.AddCounter("benchmark_counter", 1)
	}
}
