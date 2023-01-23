package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
	"github.com/go-chi/chi"
)

type MetricHandler struct {
	counterRepo repository.MetricRepository[int64]
	gaugeRepo   repository.MetricRepository[float64]
	ch          *chan int
	hashkey     string
	db          *sql.DB
}

func NewMetricHandler(app *config.AppConfig, db *sql.DB) MetricHandler {
	result := MetricHandler{
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

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")
	valueMetric := chi.URLParam(r, "value")
	switch typeMetric {
	case constants.MetricsTypeGauge:
		val, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.gaugeRepo.Add(dataclass.Metric[float64]{Name: nameMetric, Value: val})
	case constants.MetricsTypeCounter:
		val, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		_, err = h.AddCounter(nameMetric, val)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Incorrect type", http.StatusNotImplemented)
		return
	}
	h.WriteToFileIfNeeded()
	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	if !validateContentTypeIsJSON(w, r) {
		return
	}
	var model models.Metrics
	if err := decodeJSONBody(r.Body, r.Header.Get("Content-Encoding"), &model); err != nil {
		var mr *RequestError
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if h.hashkey != "" {
		hash, err := helpers.GetHash(fmt.Sprint(model), h.hashkey)
		if err != nil {
			log.Printf("Hash generation error: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		if hash != model.Hash {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}

	switch model.MType {
	case constants.MetricsTypeCounter:
		if model.Delta == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		val, err := h.AddCounter(model.ID, *model.Delta)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		model.Delta = &val
	case constants.MetricsTypeGauge:
		if model.Value == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		err := h.gaugeRepo.Add(dataclass.Metric[float64]{Name: model.ID, Value: *model.Value})
		if err != nil {
			log.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}
	h.WriteToFileIfNeeded()
	responseJSON(w, &model)
}

func (h *MetricHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	if !validateContentTypeIsJSON(w, r) {
		return
	}
	model := []models.Metrics{}
	if err := decodeJSONBody(r.Body, r.Header.Get("Content-Encoding"), &model); err != nil {
		var mr *RequestError
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	gaugeList := []dataclass.Metric[float64]{}
	counterList := []dataclass.Metric[int64]{}
	countVal := map[string]int64{}
	for _, val := range model {
		if val.MType == constants.MetricsTypeCounter {
			if _, ok := countVal[val.ID]; !ok {
				oldValue, err := h.counterRepo.GetByName(val.ID)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
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
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
	defer tx.Rollback()
	err = h.gaugeRepo.AddBatch(tx, gaugeList)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = h.counterRepo.AddBatch(tx, counterList)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tx.Commit()
	h.WriteToFileIfNeeded()
	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) AddCounter(name string, val int64) (int64, error) {
	oldValue, err := h.counterRepo.GetByName(name)
	if err != nil {
		return 0, err
	}
	if oldValue != nil {
		val += oldValue.Value
	}
	h.counterRepo.Add(dataclass.Metric[int64]{Name: name, Value: val})
	return val, nil
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")
	result := ""
	switch typeMetric {
	case constants.MetricsTypeGauge:
		met, err := h.gaugeRepo.GetByName(nameMetric)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		result = fmt.Sprintf("%.3f", met.Value)
	case constants.MetricsTypeCounter:
		met, err := h.counterRepo.GetByName(nameMetric)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		result = fmt.Sprintf("%d", met.Value)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func (h *MetricHandler) GetJSONValue(w http.ResponseWriter, r *http.Request) {
	if !validateContentTypeIsJSON(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var model models.Metrics
	if err := decodeJSONBody(r.Body, r.Header.Get("Content-Encoding"), &model); err != nil {
		var mr *RequestError
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	switch model.MType {
	case constants.MetricsTypeGauge:
		met, err := h.gaugeRepo.GetByName(model.ID)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		model.Value = &met.Value
	case constants.MetricsTypeCounter:
		met, err := h.counterRepo.GetByName(model.ID)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		model.Delta = &met.Value
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	hash, err := helpers.GetHash(fmt.Sprint(model), h.hashkey)
	if err != nil {
		log.Println(err)
	}
	model.Hash = hash
	responseJSON(w, model)
}

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	counterList, err := h.counterRepo.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	gaugeList, err := h.gaugeRepo.GetAll()
	if err != nil {
		log.Fatal(err)
	}

	result := map[string]string{}
	for _, v := range gaugeList {
		result[v.Name] = fmt.Sprintf("%.3f", v.Value)
	}
	for _, v := range counterList {
		result[v.Name] = fmt.Sprintf("%d", v.Value)
	}
	w.Header().Set("Content-Type", "text/html")
	render.Render(w, "home.html", &models.TemplateDate{Data: map[string]any{"metrics": result}})
}

func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *MetricHandler) WriteToFileIfNeeded() {
	if h.ch != nil {
		*h.ch <- 1
	}
}
