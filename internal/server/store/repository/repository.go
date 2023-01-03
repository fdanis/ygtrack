package repository

import (
	"context"

	"github.com/fdanis/ygtrack/internal/constraints"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
)

type MetricRepository[T constraints.Number] interface {
	GetAll(ctx context.Context) ([]dataclass.Metric[T], error)
	GetByName(ctx context.Context, name string) (*dataclass.Metric[T], error)
	Add(ctx context.Context, data dataclass.Metric[T]) error
}
