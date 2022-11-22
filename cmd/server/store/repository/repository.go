package repository

import (
	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
)

type GaugeRepository interface {
	GetAll() ([]dataclass.GaugeMetric, error)
	GetByName(name string) (*dataclass.GaugeMetric, error)
	Add(data dataclass.GaugeMetric) error
}

type CounterRepository interface {
	GetAll() ([]dataclass.CounterMetric, error)
	GetByName(name string) (*dataclass.CounterMetric, error)
	Add(data dataclass.CounterMetric) error
}
