package repository

import (
	"github.com/fdanis/ygtrack/internal/constraints"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
)

type MetricRepository[T constraints.Number] interface {
	GetAll() ([]dataclass.Metric[T], error)
	GetByName(name string) (*dataclass.Metric[T], error)
	Add(data dataclass.Metric[T]) error
}
