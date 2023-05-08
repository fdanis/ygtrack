package metricrepository

import (
	"database/sql"
	"errors"
	"strings"
	"sync"

	"github.com/fdanis/ygtrack/internal/constraints"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type MetricRepository[T constraints.Number] struct {
	rw          sync.RWMutex
	Datastorage map[string]dataclass.Metric[T]
}

func NewMetricRepository[T constraints.Number]() repository.MetricRepository[T] {
	return &MetricRepository[T]{
		Datastorage: map[string]dataclass.Metric[T]{},
		rw:          sync.RWMutex{},
	}
}

func (r *MetricRepository[T]) GetAll() ([]dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	list := []dataclass.Metric[T]{}
	r.rw.RLock()
	defer r.rw.RUnlock()
	for _, v := range r.Datastorage {
		list = append(list, v)
	}
	return list, nil
}

func (r *MetricRepository[T]) GetByName(name string) (*dataclass.Metric[T], error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	r.rw.RLock()
	defer r.rw.RUnlock()
	if v, ok := r.Datastorage[strings.ToLower(name)]; ok {
		return &dataclass.Metric[T]{Name: v.Name, Value: v.Value}, nil
	}
	return nil, nil
}

func (r *MetricRepository[T]) Add(data dataclass.Metric[T]) error {
	if r.Datastorage == nil {
		return errors.New("data storage is undefined")
	}
	r.rw.Lock()
	defer r.rw.Unlock()
	r.Datastorage[strings.ToLower(data.Name)] = data
	return nil
}

func (r *MetricRepository[T]) AddBatch(tx *sql.Tx, data []dataclass.Metric[T]) error {
	r.rw.Lock()
	defer r.rw.Unlock()
	for _, item := range data {
		r.Datastorage[strings.ToLower(item.Name)] = item
	}
	return nil
}
