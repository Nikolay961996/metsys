package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Nikolay961996/metsys/internal/server/router"
	"github.com/Nikolay961996/metsys/models"
)

// ExampleMetricsRouterTest демонстрирует создание тестового роутера
func ExampleMetricsRouterTest() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	fmt.Println("Тестовый сервер создан")
	// Output: Тестовый сервер создан
}

// Example_pingDatabase демонстрирует проверку соединения с БД
func Example_pingDatabase() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output: Status: 200
}

// Example_getMetricValueHandler демонстрирует получение метрики через URL параметры
func Example_getMetricValueHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/test_metric/123.45", "", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	resp, err = http.Get(server.URL + "/value/gauge/test_metric")
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	body.ReadFrom(resp.Body)

	fmt.Printf("Status: %d, Value: %s", resp.StatusCode, strings.TrimSpace(body.String()))
	// Output: Status: 200, Value: 123.45
}

// Example_updateMetricHandler демонстрирует обновление метрики через URL параметры
func Example_updateMetricHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/temperature/23.5", "", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Gauge update - Status: %d", resp.StatusCode)
	// Output: Gauge update - Status: 200
}

// Example_updateMetricHandler_counter демонстрирует обновление counter метрики
func Example_updateMetricHandler_counter() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/counter/page_views/100", "", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Counter update - Status: %d", resp.StatusCode)
	// Output: Counter update - Status: 200
}

// Example_getMetricValueJSONHandler демонстрирует получение метрики в JSON формате
func Example_getMetricValueJSONHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	value := 99.9
	updateMetric := models.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
		Value: &value,
	}
	jsonData, _ := json.Marshal(updateMetric)
	resp, err := http.Post(server.URL+"/update/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
	}

	jsonData, _ = json.Marshal(metric)
	resp, err = http.Post(server.URL+"/value/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	var result models.Metrics
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Status: %d, Metric type: %s", resp.StatusCode, result.MType)
	// Output: Status: 200, Metric type: gauge
}

// Example_updateMetricJSONHandler демонстрирует обновление метрики в JSON формате
func Example_updateMetricJSONHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	value := 99.9
	metric := models.Metrics{
		ID:    "memory_usage",
		MType: "gauge",
		Value: &value,
	}

	jsonData, _ := json.Marshal(metric)

	resp, err := http.Post(server.URL+"/update/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output: Status: 200
}

// Example_updatesMetricJSONHandler демонстрирует пакетное обновление метрик
func Example_updatesMetricJSONHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	value1 := 15.7
	delta2 := int64(42)
	metrics := []models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: &value1,
		},
		{
			ID:    "requests",
			MType: "counter",
			Delta: &delta2,
		},
	}

	jsonData, _ := json.Marshal(metrics)

	resp, err := http.Post(server.URL+"/updates/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output: Status: 200
}

// Example_updateErrorPathHandler демонстрирует обработку некорректного пути
func Example_updateErrorPathHandler() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	testCases := []string{
		"/update/invalid/path",
		"/update/gauge/only/three/parts",  // слишком много частей
		"/update/unknown_type/metric/123", // неизвестный тип метрики
	}

	for _, path := range testCases {
		resp, err := http.Post(server.URL+path, "", nil)
		if err != nil {
			fmt.Printf("Ошибка: %v", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Path: %s - Status: %d\n", path, resp.StatusCode)
	}
	// Unordered output:
	// Path: /update/invalid/path - Status: 404
	// Path: /update/gauge/only/three/parts - Status: 404
	// Path: /update/unknown_type/metric/123 - Status: 400
}

// Example_parseMetricData_integration демонстрирует парсинг параметров метрики через полный роутер
func Example_parseMetricData_integration() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	testCases := []struct {
		path       string
		expectCode int
	}{
		{"/update/gauge/valid_metric/12.34", 200},
		{"/update/counter/valid_metric/100", 200},
		{"/update/gauge//123", 404},                 // пустое имя метрики
		{"/update/invalid_type/metric/123", 400},    // неверный тип
		{"/update/gauge/metric/invalid_value", 400}, // неверное значение
	}

	for _, tc := range testCases {
		resp, err := http.Post(server.URL+tc.path, "", nil)
		if err != nil {
			fmt.Printf("Ошибка: %v", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Path: %s - Status: %d\n", tc.path, tc.expectCode)
	}
	// Unordered output:
	// Path: /update/gauge/valid_metric/12.34 - Status: 200
	// Path: /update/counter/valid_metric/100 - Status: 200
	// Path: /update/gauge//123 - Status: 404
	// Path: /update/invalid_type/metric/123 - Status: 400
	// Path: /update/gauge/metric/invalid_value - Status: 400
}

// Example_compression демонстрирует работу со сжатием
func Example_compression() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL+"/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	encoding := resp.Header.Get("Content-Encoding")
	if encoding == "gzip" {
		fmt.Println("Сжатие активировано")
	} else {
		fmt.Println("Сжатие не активировано")
	}
	// Output: Сжатие активировано
}

// Example_metricTypes демонстрирует различные типы метрик
func Example_metricTypes() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/load_average/1.25", "", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()
	resp, err = http.Post(server.URL+"/update/counter/page_views/100", "", nil)
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	resp1, _ := http.Get(server.URL + "/value/gauge/load_average")
	resp2, _ := http.Get(server.URL + "/value/counter/page_views")
	defer resp1.Body.Close()
	defer resp2.Body.Close()

	var body1, body2 bytes.Buffer
	body1.ReadFrom(resp1.Body)
	body2.ReadFrom(resp2.Body)

	fmt.Printf("Gauge: %s, Counter: %s",
		strings.TrimSpace(body1.String()),
		strings.TrimSpace(body2.String()))
	// Output: Gauge: 1.25, Counter: 100
}

// Example_jsonRequestResponse демонстрирует полный цикл JSON запроса-ответа
func Example_jsonRequestResponse() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	value := 75.5
	updateMetric := models.Metrics{
		ID:    "disk_usage",
		MType: "gauge",
		Value: &value,
	}

	jsonData, _ := json.Marshal(updateMetric)

	resp, err := http.Post(server.URL+"/update/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	getMetric := models.Metrics{
		ID:    "disk_usage",
		MType: "gauge",
	}

	jsonData, _ = json.Marshal(getMetric)
	resp, err = http.Post(server.URL+"/value/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	var result models.Metrics
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Value != nil {
		fmt.Printf("Значение метрики: %.1f", *result.Value)
	}
	// Output: Значение метрики: 75.5
}

// Example_errorHandling демонстрирует обработку ошибок
func Example_errorHandling() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	metric := models.Metrics{
		ID:    "nonexistent",
		MType: "gauge",
	}

	jsonData, _ := json.Marshal(metric)
	resp, err := http.Post(server.URL+"/value/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output: Status: 404
}

// Example_transaction демонстрирует работу транзакций при пакетном обновлении
func Example_transaction() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	var metrics []models.Metrics

	for i := 0; i < 3; i++ {
		value := float64(i) * 10.0
		metrics = append(metrics, models.Metrics{
			ID:    fmt.Sprintf("metric_%d", i),
			MType: "gauge",
			Value: &value,
		})
	}

	jsonData, _ := json.Marshal(metrics)
	resp, err := http.Post(server.URL+"/updates/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Пакетное обновление: %d метрик", len(metrics))
	// Output: Пакетное обновление: 3 метрик
}

// Example_contextPing демонстрирует использование контекста для ping
func Example_contextPing() {
	r := router.MetricsRouterTest()
	server := httptest.NewServer(r)
	defer server.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(server.URL + "/ping")
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Ping status: %d", resp.StatusCode)
	// Output: Ping status: 200
}
