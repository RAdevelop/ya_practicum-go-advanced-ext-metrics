// Package models - представление метрики.
package models

import "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"

// Поддерживаемые типы метрик.
const (
	// Counter — счётчик: значение типа int64, накапливается (суммируется).
	Counter = "counter"
	// Gauge — измеритель: значение типа float64, перезаписывается при каждом обновлении.
	Gauge = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Ограничиваемся плоской моделью.
//
// Metrics — модель метрики в формате обмена с агентом.
// Delta и Value объявлены указателями, чтобы отличать заданное значение "0"
// от незаданного и не кодировать пустые поля в JSON.
type Metrics struct {
	ID    string   `json:"id" db:"metric_id"`
	MType string   `json:"type" db:"m_type"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"hash,omitempty" db:"hash"`
}

// StrValue — возвращает значение метрики в строковом виде
// в зависимости от её типа. Для gauge это Value, для counter — Delta.
// Возвращает пустую строку, если метрика пустая или значение не задано.
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
