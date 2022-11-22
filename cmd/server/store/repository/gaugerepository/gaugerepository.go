package gaugerepository

import (
	"errors"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
)

type GaugeRepository struct {
	Datastorage *map[string]float64
}

func NewGaugeRepository() GaugeRepository {
	return GaugeRepository{Datastorage: &map[string]float64{}}
}

func (r *GaugeRepository) GetAll() ([]dataclass.GaugeMetric, error) {
	if r.Datastorage == nil {
		return nil, errors.New("Data storage is not difened")
	}
	list := []dataclass.GaugeMetric{}
	for k, v := range *r.Datastorage {
		list = append(list, dataclass.GaugeMetric{Name: k, Value: v})
	}
	return list, nil
}
func (r *GaugeRepository) GetByName(name string) (*dataclass.GaugeMetric, error) {
	if r.Datastorage == nil {
		return nil, errors.New("Data storage is not difened")
	}
	if v, ok := (*r.Datastorage)[name]; ok {
		return &dataclass.GaugeMetric{Name: name, Value: v}, nil
	}
	return nil, nil
}

func (r *GaugeRepository) Add(data dataclass.GaugeMetric) error {
	if r.Datastorage == nil {
		return errors.New("Data storage is not difened")
	}
	(*r.Datastorage)[data.Name] = data.Value
	return nil
}
