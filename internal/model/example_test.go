package models_test

import (
	"encoding/json"
	"fmt"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
)

// Example демонстрирует базовое использование структуры Metrics.
// Показывает создание и проверку валидности метрики типа gauge.
func Example() {
	// Создаем метрику gauge
	value := 42.5
	metric := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: &value,
	}

	// Проверяем валидность
	fmt.Println(metric.IsValid())
	// Output:
	// true
}

// ExampleMetrics_IsValid демонстрирует проверку валидности метрики.
// Показывает, как проверить корректность структуры метрики.
func ExampleMetrics_IsValid() {
	value := 100.0
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: models.Gauge,
		Value: &value,
	}

	fmt.Println(metric.IsValid())
	// Output:
	// true
}

// ExampleMetrics_IsValid_invalid демонстрирует невалидную метрику.
// Показывает случай, когда метрика gauge не имеет значения.
func ExampleMetrics_IsValid_invalid() {
	// Метрика gauge без значения невалидна
	metric := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: nil, // отсутствует значение
	}

	fmt.Println(metric.IsValid())
	// Output:
	// false
}

// ExampleMetrics_GetGaugeValue демонстрирует получение значения gauge метрики.
// Показывает, как безопасно извлечь значение из метрики типа gauge.
func ExampleMetrics_GetGaugeValue() {
	value := 75.5
	metric := models.Metrics{
		ID:    "memory_usage",
		MType: models.Gauge,
		Value: &value,
	}

	val, ok := metric.GetGaugeValue()
	if ok {
		fmt.Printf("Value: %.1f\n", val)
	}
	// Output:
	// Value: 75.5
}

// ExampleMetrics_GetCounterValue демонстрирует получение значения counter метрики.
// Показывает, как безопасно извлечь значение из метрики типа counter.
func ExampleMetrics_GetCounterValue() {
	delta := int64(42)
	metric := models.Metrics{
		ID:    "request_count",
		MType: models.Counter,
		Delta: &delta,
	}

	val, ok := metric.GetCounterValue()
	if ok {
		fmt.Printf("Delta: %d\n", val)
	}
	// Output:
	// Delta: 42
}

// ExampleMetrics_gauge демонстрирует создание метрики типа gauge.
// Показывает правильную инициализацию gauge метрики.
func ExampleMetrics_gauge() {
	value := 23.45
	metric := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: &value,
	}

	jsonData, _ := json.Marshal(metric)
	fmt.Println(string(jsonData))
	// Output:
	// {"id":"temperature","type":"gauge","value":23.45}
}

// ExampleMetrics_counter демонстрирует создание метрики типа counter.
// Показывает правильную инициализацию counter метрики.
func ExampleMetrics_counter() {
	delta := int64(10)
	metric := models.Metrics{
		ID:    "requests",
		MType: models.Counter,
		Delta: &delta,
	}

	jsonData, _ := json.Marshal(metric)
	fmt.Println(string(jsonData))
	// Output:
	// {"id":"requests","type":"counter","delta":10}
}

// ExampleCounter демонстрирует использование константы типа counter.
// Показывает, как использовать предопределенный тип метрики.
func ExampleCounter() {
	fmt.Println(models.Counter)
	// Output:
	// counter
}

// ExampleGauge демонстрирует использование константы типа gauge.
// Показывает, как использовать предопределенный тип метрики.
func ExampleGauge() {
	fmt.Println(models.Gauge)
	// Output:
	// gauge
}
