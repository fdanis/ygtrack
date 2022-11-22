package counterrepository

import (
	"errors"
	"fmt"

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
		return nil, errors.New("Data storage is not difened")
	}
	list := []dataclass.CounterMetric{}
	for k, v := range *r.Datastorage {
		list = append(list, dataclass.CounterMetric{Name: k, Value: v})
	}
	return list, nil
}

func (r *CounterRepository) GetByName(name string) (*dataclass.CounterMetric, error) {
	if r.Datastorage == nil {
		return nil, errors.New("Data storage is not difened")
	}
	fmt.Println(*r.Datastorage)
	fmt.Println("KEyName" + name)
	if v, ok := (*r.Datastorage)[name]; ok {

		fmt.Println("GotValue")
		return &dataclass.CounterMetric{Name: name, Value: v}, nil
	}
	fmt.Println("empty")
	return nil, nil
}

func (r *CounterRepository) Add(data dataclass.CounterMetric) error {
	if r.Datastorage == nil {
		return errors.New("Data storage is not difened")
	}
	(*r.Datastorage)[data.Name] += data.Value
	return nil
}
