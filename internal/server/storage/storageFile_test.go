package storage_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Nikolay961996/metsys/internal/server/storage"
)

// TestFileStorage_BasicOperations тестирует базовые операции FileStorage
func TestFileStorage_BasicOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	s := storage.NewFileStorage(tmpFile.Name(), 0, false) // sync mode
	defer s.Close()

	s.SetGauge("temperature", 23.5)
	value, err := s.GetGauge("temperature")
	if err != nil {
		t.Errorf("GetGauge failed: %v", err)
	}
	if value != 23.5 {
		t.Errorf("Expected 23.5, got %f", value)
	}

	s.AddCounter("requests", 10)
	s.AddCounter("requests", 5)
	valueInt, err := s.GetCounter("requests")
	if err != nil {
		t.Errorf("GetCounter failed: %v", err)
	}
	if valueInt != 15 {
		t.Errorf("Expected 15, got %d", valueInt)
	}
}

// TestFileStorage_Persistence тестирует сохранение и восстановление данных
func TestFileStorage_Persistence(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_persistence_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	s1 := storage.NewFileStorage(tmpFile.Name(), 0, false)
	s1.SetGauge("cpu_usage", 75.5)
	s1.AddCounter("requests", 100)
	s1.Close()

	s2 := storage.NewFileStorage(tmpFile.Name(), 0, true)
	defer s2.Close()

	value, err := s2.GetGauge("cpu_usage")
	if err != nil {
		t.Errorf("Failed to restore gauge: %v", err)
	}
	if value != 75.5 {
		t.Errorf("Expected gauge value 75.5, got %f", value)
	}

	valueInt, err := s2.GetCounter("requests")
	if err != nil {
		t.Errorf("Failed to restore counter: %v", err)
	}
	if valueInt != 100 {
		t.Errorf("Expected counter value 100, got %d", valueInt)
	}
}

// TestFileStorage_AsyncSave тестирует асинхронное сохранение
func TestFileStorage_AsyncSave(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_async_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	savePeriod := 10 * time.Millisecond
	s := storage.NewFileStorage(tmpFile.Name(), savePeriod, false)
	defer s.Close()

	s.SetGauge("async_metric", 99.9)
	s.AddCounter("async_counter", 42)

	time.Sleep(2 * savePeriod)

	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Errorf("Failed to stat file: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Error("File should not be empty after async save")
	}
}

// TestFileStorage_GetAll тестирует получение всех метрик
func TestFileStorage_GetAll(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_getall_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	s := storage.NewFileStorage(tmpFile.Name(), 0, false)
	defer s.Close()

	s.SetGauge("metric1", 1.0)
	s.SetGauge("metric2", 2.0)
	s.AddCounter("counter1", 10)
	s.AddCounter("counter2", 20)

	metrics := s.GetAll()
	if len(metrics) != 4 {
		t.Errorf("Expected 4 metrics, got %d", len(metrics))
	}
}

// TestFileStorage_ErrorHandling тестирует обработку ошибок
func TestFileStorage_ErrorHandling(t *testing.T) {
	invalidPath := "/invalid/path/storage.tmp"
	s := storage.NewFileStorage(invalidPath, 0, false)
	defer s.Close()

	s.SetGauge("test", 123.45)
	value, err := s.GetGauge("test")
	if err != nil {
		t.Errorf("Operations should work even with invalid file path: %v", err)
	}
	if value != 123.45 {
		t.Errorf("Expected 123.45, got %f", value)
	}
}

// TestFileStorage_RestoreNonExistentFile тестирует восстановление из несуществующего файла
func TestFileStorage_RestoreNonExistentFile(t *testing.T) {
	s := storage.NewFileStorage("/nonexistent/file.tmp", 0, true)
	defer s.Close()

	metrics := s.GetAll()
	if len(metrics) != 0 {
		t.Errorf("Expected empty storage, got %d metrics", len(metrics))
	}
}

// TestFileStorage_CloseSafety тестирует безопасное закрытие
func TestFileStorage_CloseSafety(t *testing.T) {
	tmpFile1, err := os.CreateTemp("", "test_close1_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile1.Close()
	defer os.Remove(tmpFile1.Name())

	s1 := storage.NewFileStorage(tmpFile1.Name(), 0, false)
	s1.Close()

	tmpFile2, err := os.CreateTemp("", "test_close2_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile2.Close()
	defer os.Remove(tmpFile2.Name())

	s2 := storage.NewFileStorage(tmpFile2.Name(), time.Minute, false) // async mode
	s2.Close()

	s3 := storage.NewFileStorage(tmpFile1.Name(), 0, false)
	s3.Close()
	s3.Close() // Should not panic on double close
}

// ExampleFileStorage демонстрирует базовое использование FileStorage
func ExampleFileStorage() {
	tmpFile, err := os.CreateTemp("", "example_*.tmp")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage := storage.NewFileStorage(tmpFile.Name(), 0, false)
	defer storage.Close()

	storage.SetGauge("temperature", 23.5)
	storage.AddCounter("requests", 100)

	value, _ := storage.GetGauge("temperature")
	count, _ := storage.GetCounter("requests")

	fmt.Printf("Temperature: %.1f, Requests: %d\n", value, count)

	// Output:
	// Temperature: 23.5, Requests: 100
}

// ExampleFileStorage_persistence демонстрирует сохранение и восстановление
func ExampleFileStorage_persistence() {
	tmpFile, err := os.CreateTemp("", "persistence_*.tmp")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage1 := storage.NewFileStorage(tmpFile.Name(), 0, false)
	storage1.SetGauge("cpu", 75.5)
	storage1.AddCounter("hits", 42)
	storage1.Close()

	storage2 := storage.NewFileStorage(tmpFile.Name(), 0, true)
	defer storage2.Close()

	cpu, _ := storage2.GetGauge("cpu")
	hits, _ := storage2.GetCounter("hits")

	fmt.Printf("CPU: %.1f%%, Hits: %d\n", cpu, hits)

	// Output:
	// CPU: 75.5%, Hits: 42
}

// ExampleFileStorage_async демонстрирует асинхронное сохранение
func ExampleFileStorage_async() {
	tmpFile, err := os.CreateTemp("", "async_*.tmp")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage := storage.NewFileStorage(tmpFile.Name(), 100*time.Millisecond, false)
	defer storage.Close()

	storage.SetGauge("memory", 65.2)
	storage.AddCounter("visitors", 1000)

	// Даем время для автосохранения
	time.Sleep(150 * time.Millisecond)

	fmt.Println("Data saved asynchronously")

	// Output:
	// Data saved asynchronously
}

// BenchmarkFileStorage_SetGauge бенчмарк для SetGauge в FileStorage
func BenchmarkFileStorage_SetGauge(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchmark_*.tmp")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	s := storage.NewFileStorage(tmpFile.Name(), 0, false)
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetGauge("benchmark_metric", float64(i))
	}
}

// BenchmarkFileStorage_AddCounter бенчмарк для AddCounter в FileStorage
func BenchmarkFileStorage_AddCounter(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchmark_*.tmp")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	s := storage.NewFileStorage(tmpFile.Name(), 0, false)
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.AddCounter("benchmark_counter", 1)
	}
}
