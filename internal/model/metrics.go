// Package models содержит структуры данных для работы с метриками.
package models

const (
	// Counter представляет тип метрики счетчик (накопительное значение).
	Counter = "counter"

	// Gauge представляет тип метрики измерение (текущее значение).
	Gauge = "gauge"
)

// Metrics представляет структуру метрики для передачи между клиентом и сервером.
// Поля Delta и Value объявлены через указатели для различения нулевого значения
// и отсутствующего значения, что позволяет корректно обрабатывать JSON.
type Metrics struct {
	ID    string   `json:"id"`              // ID - уникальный идентификатор метрики
	MType string   `json:"type"`            // MType - тип метрики (counter или gauge)
	Delta *int64   `json:"delta,omitempty"` // Delta - значение для метрики типа counter
	Value *float64 `json:"value,omitempty"` // Value - значение для метрики типа gauge
	Hash  string   `json:"hash,omitempty"`  // Hash - HMAC хеш для проверки целостности данных

}

// IsValid проверяет корректность структуры метрики.
// Возвращает true, если метрика имеет непустой ID и соответствующее типу значение.
func (M *Metrics) IsValid() bool {
	if M.ID == "" {
		return false
	}
	switch M.MType {
	case Gauge:
		return M.Value != nil
	case Counter:
		return M.Delta != nil
	default:
		return false
	}
}

// GetGaugeValue возвращает значение метрики типа gauge.
// Возвращает значение и true, если метрика валидна и имеет тип gauge.
// В противном случае возвращает 0 и false.
func (M *Metrics) GetGaugeValue() (float64, bool) {
	if !M.IsValid() {
		return 0, false
	}
	return *M.Value, true
}

// GetCounterValue возвращает значение метрики типа counter.
// Возвращает значение и true, если метрика валидна и имеет тип counter.
// В противном случае возвращает 0 и false.
func (M *Metrics) GetCounterValue() (int64, bool) {
	if !M.IsValid() {
		return 0, false
	}
	return *M.Delta, true
}
