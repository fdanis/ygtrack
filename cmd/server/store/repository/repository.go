package repository

import (
	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
)

type MetricRepository[T any] interface {
	GetAll() ([]dataclass.Metric[T], error)
	GetByName(name string) (*dataclass.Metric[T], error)
	Add(data dataclass.Metric[T]) error
}
