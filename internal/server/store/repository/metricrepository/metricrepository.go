package metricrepository

import (
	"context"
	"errors"
	"strings"

	"github.com/fdanis/ygtrack/internal/constraints"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type MetricRepository[T constraints.Number] struct {
	Datastorage *map[string]dataclass.Metric[T]
}

func NewMetricRepository[T constraints.Number]() repository.MetricRepository[T] {
	return MetricRepository[T]{Datastorage: &map[string]dataclass.Metric[T]{}}
}

func (r MetricRepository[T]) GetAll(ctx context.Context) ([]dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	list := []dataclass.Metric[T]{}
	for _, v := range *r.Datastorage {
		list = append(list, v)
	}
	return list, nil
}

func (r MetricRepository[T]) GetByName(ctx context.Context, name string) (*dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	if v, ok := (*r.Datastorage)[strings.ToLower(name)]; ok {
		return &dataclass.Metric[T]{Name: v.Name, Value: v.Value}, nil
	}
	return nil, nil
}

func (r MetricRepository[T]) Add(ctx context.Context, data dataclass.Metric[T]) error {
	if r.Datastorage == nil {
		return errors.New("data storage is undefined")
	}
	(*r.Datastorage)[strings.ToLower(data.Name)] = data
	return nil
}
