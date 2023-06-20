package metricsservice

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type MetricsService struct {
	ch          *chan int
	counterRepo repository.MetricRepository[int64]
	gaugeRepo   repository.MetricRepository[float64]
	hashkey     string
	db          *sql.DB
}

func NewMetricsService(app *config.AppConfig, db *sql.DB) *MetricsService {
	result := &MetricsService{
		counterRepo: app.CounterRepository,
		gaugeRepo:   app.GaugeRepository,
		hashkey:     app.Parameters.Key,
		db:          db,
	}
	if app.SaveToFileSync {
		result.ch = &app.ChForSyncWithFile
	}
	return result
}

func (m *MetricsService) AddMetric(model models.Metrics) error {

	if m.hashkey != "" {
		hash, err := helpers.GetHash(model, m.hashkey)
		if err != nil {
			log.Printf("Hash generation error: %v", err)
			return err
		}
		if hash != model.Hash {
			return NewMetricsError(0, fmt.Sprintf("hash != model.Hash; %s != %s; %#v", hash, model.Hash, model))
		}
	}

	switch model.MType {
	case constants.MetricsTypeCounter:
		if model.Delta == nil {
			return NewMetricsError(0, fmt.Sprintf("model.Delta == nil; %v", model))
		}
		val, err := m.addCounter(model.ID, *model.Delta)
		if err != nil {
			log.Printf("Add counter: %v", err)
			return err
		}
		model.Delta = &val
	case constants.MetricsTypeGauge:
		if model.Value == nil {
			return NewMetricsError(0, fmt.Sprintf("model.Value == nil; %v", model))
		}
		err := m.gaugeRepo.Add(dataclass.Metric[float64]{Name: model.ID, Value: *model.Value})
		if err != nil {
			log.Println(err)
			return err
		}
	default:
		return fmt.Errorf("not implemented")
	}
	go m.writeToFileIfNeeded()
	return nil
}

func (m *MetricsService) UpdateBatch(model []models.Metrics) error {
	gaugeList := []dataclass.Metric[float64]{}
	counterList := []dataclass.Metric[int64]{}
	countVal := map[string]int64{}
	for _, val := range model {
		if val.MType == constants.MetricsTypeCounter {
			if _, ok := countVal[val.ID]; !ok {
				oldValue, err := m.counterRepo.GetByName(val.ID)
				if err != nil {
					log.Println(err)
					return err
				}
				if oldValue != nil {
					countVal[val.ID] = oldValue.Value
				} else {
					countVal[val.ID] = 0
				}
			}
			countVal[val.ID] += *val.Delta
			counterList = append(counterList, dataclass.Metric[int64]{Name: val.ID, Value: countVal[val.ID]})

		} else if val.MType == constants.MetricsTypeGauge {
			gaugeList = append(gaugeList, dataclass.Metric[float64]{Name: val.ID, Value: *val.Value})
		} else {
			return NewMetricsError(0, "incorect type")
		}
	}

	tx, err := m.db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()
	err = m.gaugeRepo.AddBatch(tx, gaugeList)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	err = m.counterRepo.AddBatch(tx, counterList)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	tx.Commit()
	go m.writeToFileIfNeeded()
	return nil
}

func (m *MetricsService) GetCounterValue(name string) (int64, error) {
	met, err := m.counterRepo.GetByName(name)
	if err != nil {
		log.Print(err)
	}
	if met == nil {
		return 0, NewMetricsError(1, "Not found")
	}
	return met.Value, nil
}

func (m *MetricsService) GetGaugeValue(name string) (float64, error) {
	met, err := m.gaugeRepo.GetByName(name)
	if err != nil {
		log.Print(err)
	}
	if met == nil {
		return 0, NewMetricsError(1, "Not found")
	}
	return met.Value, nil
}

func (m *MetricsService) GetAllCounter() ([]dataclass.Metric[int64], error) {
	counterList, err := m.counterRepo.GetAll()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return counterList, nil
}

func (m *MetricsService) GetAllGauge() ([]dataclass.Metric[float64], error) {
	counterList, err := m.gaugeRepo.GetAll()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return counterList, nil
}

func (m *MetricsService) writeToFileIfNeeded() {
	if m.ch != nil {
		*m.ch <- 1
	}
}

func (m *MetricsService) addCounter(name string, val int64) (int64, error) {
	oldValue, err := m.counterRepo.GetByName(name)
	if err != nil {
		return 0, err
	}
	if oldValue != nil {
		val += oldValue.Value
	}
	m.counterRepo.Add(dataclass.Metric[int64]{Name: name, Value: val})
	return val, nil
}
