package repository_test

import (
	"fmt"

	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

// Example демонстрирует базовое использование хранилища метрик.
// Показывает создание хранилища и работу с метриками.
func Example() {
	// Создаем хранилище в памяти
	storage := repository.NewMemStorage()

	// Сохраняем gauge метрику
	storage.UpdateGauge("temperature", 23.5)

	// Получаем значение
	value, exists := storage.GetGauge("temperature")
	if exists {
		fmt.Printf("Temperature: %.1f\n", value)
	}
	// Output:
	// Temperature: 23.5
}

// ExampleNewMemStorage демонстрирует создание нового хранилища в памяти.
// Показывает инициализацию пустого хранилища метрик.
func ExampleNewMemStorage() {
	storage := repository.NewMemStorage()
	fmt.Printf("Storage created: %T\n", storage)
	// Output:
	// Storage created: *repository.MemStorage
}

// ExampleMemStorage_UpdateGauge демонстрирует обновление gauge метрики.
// Показывает, как сохранить значение метрики типа gauge.
func ExampleMemStorage_UpdateGauge() {
	storage := repository.NewMemStorage()

	// Обновляем gauge метрику
	err := storage.UpdateGauge("cpu_usage", 75.5)
	if err == nil {
		fmt.Println("Gauge updated successfully")
	}
	// Output:
	// Gauge updated successfully
}

// ExampleMemStorage_UpdateCounter демонстрирует обновление counter метрики.
// Показывает, как увеличить значение счетчика.
func ExampleMemStorage_UpdateCounter() {
	storage := repository.NewMemStorage()

	// Первое обновление
	storage.UpdateCounter("requests", 10)

	// Второе обновление (добавляется к предыдущему)
	storage.UpdateCounter("requests", 5)

	// Получаем итоговое значение
	value, _ := storage.GetCounter("requests")
	fmt.Printf("Total requests: %d\n", value)
	// Output:
	// Total requests: 15
}

// ExampleMemStorage_GetGauge демонстрирует получение gauge метрики.
// Показывает, как безопасно получить значение метрики.
func ExampleMemStorage_GetGauge() {
	storage := repository.NewMemStorage()
	storage.UpdateGauge("temperature", 22.3)

	value, exists := storage.GetGauge("temperature")
	if exists {
		fmt.Printf("Value: %.1f\n", value)
	}
	// Output:
	// Value: 22.3
}

// ExampleMemStorage_GetGauge_notFound демонстрирует поиск несуществующей метрики.
// Показывает обработку случая, когда метрика не найдена.
func ExampleMemStorage_GetGauge_notFound() {
	storage := repository.NewMemStorage()

	_, exists := storage.GetGauge("nonexistent")
	fmt.Printf("Exists: %t\n", exists)
	// Output:
	// Exists: false
}

// ExampleMemStorage_GetCounter демонстрирует получение counter метрики.
// Показывает, как получить значение счетчика.
func ExampleMemStorage_GetCounter() {
	storage := repository.NewMemStorage()
	storage.UpdateCounter("page_views", 100)

	value, exists := storage.GetCounter("page_views")
	if exists {
		fmt.Printf("Page views: %d\n", value)
	}
	// Output:
	// Page views: 100
}

// ExampleMemStorage_GetMetric демонстрирует получение метрики любого типа.
// Показывает универсальный способ получения метрики.
func ExampleMemStorage_GetMetric() {
	storage := repository.NewMemStorage()
	storage.UpdateGauge("temperature", 25.5)

	metric, exists := storage.GetMetric("temperature")
	if exists {
		fmt.Printf("Type: %s, Value: %.1f\n", metric.MType, *metric.Value)
	}
	// Output:
	// Type: gauge, Value: 25.5
}

// ExampleMemStorage_GetAllMetrics демонстрирует получение всех метрик.
// Показывает, как получить список всех сохраненных метрик.
func ExampleMemStorage_GetAllMetrics() {
	storage := repository.NewMemStorage()

	// Добавляем несколько метрик
	storage.UpdateGauge("cpu", 45.5)
	storage.UpdateCounter("requests", 10)

	metrics := storage.GetAllMetrics()
	fmt.Printf("Total metrics: %d\n", len(metrics))
	// Output:
	// Total metrics: 2
}

// ExampleMemStorage_concurrentUpdates демонстрирует безопасность при конкурентном доступе.
// Показывает, что хранилище можно использовать из нескольких горутин.
func ExampleMemStorage_concurrentUpdates() {
	storage := repository.NewMemStorage()

	// Обновляем метрику несколько раз
	storage.UpdateCounter("concurrent", 1)
	storage.UpdateCounter("concurrent", 2)
	storage.UpdateCounter("concurrent", 3)

	value, _ := storage.GetCounter("concurrent")
	fmt.Printf("Total: %d\n", value)
	// Output:
	// Total: 6
}

// ExampleStorage интерфейс демонстрирует использование интерфейса Storage.
// Показывает, что MemStorage реализует интерфейс Storage.
func ExampleStorage() {
	// Storage это интерфейс, который объединяет чтение и запись
	var storage repository.Storage
	storage = repository.NewMemStorage()

	// Используем методы интерфейса
	storage.UpdateGauge("test", 42.0)
	value, _ := storage.GetGauge("test")

	fmt.Printf("Value: %.0f\n", value)
	// Output:
	// Value: 42
}

// ExampleMetricReader демонстрирует использование интерфейса MetricReader.
// Показывает работу с метриками только для чтения.
func ExampleMetricReader() {
	storage := repository.NewMemStorage()
	storage.UpdateGauge("readonly", 99.9)

	// Используем только методы чтения
	var reader repository.MetricReader = storage
	value, exists := reader.GetGauge("readonly")

	fmt.Printf("Exists: %t, Value: %.1f\n", exists, value)
	// Output:
	// Exists: true, Value: 99.9
}

// ExampleMetricWriter демонстрирует использование интерфейса MetricWriter.
// Показывает работу с метриками только для записи.
func ExampleMetricWriter() {
	storage := repository.NewMemStorage()

	// Используем только методы записи
	var writer repository.MetricWriter = storage
	err := writer.UpdateGauge("writeonly", 123.45)

	if err == nil {
		fmt.Println("Write successful")
	}
	// Output:
	// Write successful
}
