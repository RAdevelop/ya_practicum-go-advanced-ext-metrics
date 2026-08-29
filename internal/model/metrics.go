package models

import "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"

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
	ID    string   `json:"id" db:"metric_id"`
	MType string   `json:"type" db:"m_type"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) StrValue() string {
	if m == nil {
		return ""
	}

	switch m.MType {
	case Gauge:
		if m.Value != nil {
			return converter.NumericToString(*m.Value)
		}
	case Counter:
		if m.Delta != nil {
			return converter.NumericToString(*m.Delta)
		}
	}
	return ""
}
