package models

import "testing"

func TestMetrics_Reset(t *testing.T) {
	delta := int64(42)
	value := float64(3.14)
	m := &Metrics{
		ID:    "test",
		MType: "gauge",
		Delta: &delta,
		Value: &value,
		Hash:  "hash",
	}

	m.Reset()

	if m.ID != "" {
		t.Errorf("ID not reset: %q", m.ID)
	}
	if m.MType != "" {
		t.Errorf("MType not reset: %q", m.MType)
	}
	if m.Delta == nil || *m.Delta != 0 {
		t.Errorf("Delta not reset: %v", m.Delta)
	}
	if m.Value == nil || *m.Value != 0 {
		t.Errorf("Value not reset: %v", m.Value)
	}
	if m.Hash != "" {
		t.Errorf("Hash not reset: %q", m.Hash)
	}
}

func TestMetrics_Reset_Nil(t *testing.T) {
	var m *Metrics
	m.Reset() // не должно паниковать
}
