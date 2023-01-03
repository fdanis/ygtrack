package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
}

func NewMetricHandler(app *config.AppConfig) MetricHandler {
	result := MetricHandler{
		counterRepo: app.CounterRepository,
		gaugeRepo:   app.GaugeRepository,
		hashkey:     app.Parameters.Key,
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	switch typeMetric {
	case "gauge":
		val, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.gaugeRepo.Add(ctx, dataclass.Metric[float64]{Name: nameMetric, Value: val})
	case "counter":
		val, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		_, err = h.AddCounter(ctx, nameMetric, val)
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
		oldHash := model.Hash
		if err := model.RefreshHash(h.hashkey); err != nil {
			log.Printf("Hash generation error: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		if oldHash != model.Hash {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}

	switch model.MType {
	case "counter":
		if model.Delta == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		val, err := h.AddCounter(ctx, model.ID, *model.Delta)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		model.Delta = &val
	case "gauge":
		if model.Value == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		err := h.gaugeRepo.Add(ctx, dataclass.Metric[float64]{Name: model.ID, Value: *model.Value})
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}
	h.WriteToFileIfNeeded()
	responseJSON(w, &model)
}

func (h *MetricHandler) AddCounter(ctx context.Context, name string, val int64) (int64, error) {

	ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	oldValue, err := h.counterRepo.GetByName(ctx1, name)
	if err != nil {
		return 0, err
	}
	if oldValue != nil {
		val += oldValue.Value
	}
	ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	h.counterRepo.Add(ctx2, dataclass.Metric[int64]{Name: name, Value: val})
	return val, nil
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")
	result := ""
	switch typeMetric {
	case "gauge":
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		met, err := h.gaugeRepo.GetByName(ctx, nameMetric)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		result = fmt.Sprintf("%.3f", met.Value)
	case "counter":
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		met, err := h.counterRepo.GetByName(ctx, nameMetric)
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
	case "gauge":
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		met, err := h.gaugeRepo.GetByName(ctx, model.ID)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		model.Value = &met.Value
	case "counter":
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		met, err := h.counterRepo.GetByName(ctx, model.ID)
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
	model.RefreshHash(h.hashkey)
	responseJSON(w, model)
}

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx1, cancel1 := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel1()
	counterList, err := h.counterRepo.GetAll(ctx1)
	if err != nil {
		log.Fatal(err)
	}
	ctx2, cancel2 := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel2()
	gaugeList, err := h.gaugeRepo.GetAll(ctx2)
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

func (h *MetricHandler) WriteToFileIfNeeded() {
	if h.ch != nil {
		*h.ch <- 1
	}
}
