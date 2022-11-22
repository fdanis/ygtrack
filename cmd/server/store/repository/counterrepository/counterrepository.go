package counterrepository

import (
	"errors"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
)

type CounterRepository struct {
	Datastorage *map[string]uint64
}

func NewCounterRepository() CounterRepository {
	return CounterRepository{Datastorage: &map[string]uint64{}}
}

func (r *CounterRepository) GetAll() ([]dataclass.CounterMetric, error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	list := []dataclass.CounterMetric{}
	for k, v := range *r.Datastorage {
		list = append(list, dataclass.CounterMetric{Name: k, Value: v})
	}
	return list, nil
}

func (r *CounterRepository) GetByName(name string) (*dataclass.CounterMetric, error) {
	if r.Datastorage == nil {
		return nil, errors.New("data storage is undefined")
	}
	if v, ok := (*r.Datastorage)[name]; ok {
		return &dataclass.CounterMetric{Name: name, Value: v}, nil
	}
	return nil, nil
}

func (r *CounterRepository) Add(data dataclass.CounterMetric) error {
	if r.Datastorage == nil {
		return errors.New("data storage is undefined")
	}
	(*r.Datastorage)[data.Name] += data.Value
	return nil
}
