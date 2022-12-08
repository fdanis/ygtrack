package metricrepository

import (
	"errors"
	"strings"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/constraints"
)

type MetricRepository[T constraints.Number] struct {
	Datastorage *map[string]dataclass.Metric[T]
}

func NewMetricRepository[T constraints.Number]() MetricRepository[T] {
	return MetricRepository[T]{Datastorage: &map[string]dataclass.Metric[T]{}}
}

func (r *MetricRepository[T]) GetAll() ([]dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	list := []dataclass.Metric[T]{}
	for _, v := range *r.Datastorage {
		list = append(list, v)
	}
	return list, nil
}

func (r *MetricRepository[T]) GetByName(name string) (*dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	if v, ok := (*r.Datastorage)[strings.ToLower(name)]; ok {
		return &dataclass.Metric[T]{Name: v.Name, Value: v.Value}, nil
	}
	return nil, nil
}

func (r *MetricRepository[T]) Add(data dataclass.Metric[T]) error {
	if r.Datastorage == nil {
		return errors.New("data storage is undefined")
	}
	(*r.Datastorage)[strings.ToLower(data.Name)] = data
	return nil
}
