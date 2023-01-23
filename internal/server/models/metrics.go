package models

import (
	"encoding/json"
	"fmt"

	"github.com/fdanis/ygtrack/internal/constants"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
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
	case constants.MetricsTypeGauge:
		var value float64
		if aliasValue.Value != nil {
			if err := json.Unmarshal(aliasValue.Value, &value); err != nil {
				return err
			}
			m.Value = &value
		}
	case constants.MetricsTypeCounter:
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

func (m *Metrics) SetHash(hash string) {
	m.Hash = hash
}

func (m Metrics) String() string {
	if m.MType == constants.MetricsTypeCounter {
		return fmt.Sprintf("%s:%s:%d", m.ID, constants.MetricsTypeCounter, *m.Delta)
	}
	return fmt.Sprintf("%s:%s:%f", m.ID, constants.MetricsTypeGauge, *m.Value)
}
