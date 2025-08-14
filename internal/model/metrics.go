package models

const (
	Counter = "counter"
	Gauge   = "gauge"
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

func (M *Metrics) IsValid() bool {
	if M.ID == "" {
		return false
	}
	switch M.MType {
	case Gauge:
		return M.Value != nil && M.Delta != nil
	case Counter:
		return M.Delta != nil && M.Value != nil
	default:
		return false
	}
}

func (M *Metrics) GetGuageValue() (float64, bool) {
	if M.MType != Gauge || M.Value == nil {
		return 0, false
	}
	return *M.Value, true
}

func (M *Metrics) GetCounterValue() (int64, bool) {
	if M.MType != Counter || M.Delta == nil {
		return 0, false
	}
	return *M.Delta, true
}
