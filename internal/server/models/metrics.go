package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m *Metrics) RefreshHash(key string) error {
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		if m.MType == "counter" {
			if _, err := h.Write([]byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta))); err != nil {
				log.Println("can not get hash for counter")
				return err
			}
		} else {
			if _, err := h.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value))); err != nil {
				log.Println("can not get hash for gauge")
				return err
			}
		}
		m.Hash = string(h.Sum(nil))
	}
	return nil
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
