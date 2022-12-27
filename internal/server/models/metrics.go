package models

import (
	"encoding/json"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) UnmarshalJSON(data []byte) error {
	type MetricsAlias Metrics

	aliasValue := &struct {
		*MetricsAlias
		Delta json.RawMessage `json:"delta,omitempty"`
		Value json.RawMessage `json:"value,omitempty"`
	}{
		MetricsAlias: (*MetricsAlias)(m),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	switch m.MType {
	case "gauge":
		var value float64
		if aliasValue.Value != nil {
			if err := json.Unmarshal(aliasValue.Value, &value); err != nil {
				return err
			}
			m.Value = &value
		}
	case "counter":
		var delta int64
		if aliasValue.Delta != nil {
			if err := json.Unmarshal(aliasValue.Delta, &delta); err != nil {
				return err
			}
			m.Delta = &delta
		}
	}
	return nil
}
