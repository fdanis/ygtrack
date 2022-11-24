package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fdanis/ygtrack/cmd/server/models"
	"github.com/fdanis/ygtrack/cmd/server/render"
	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
	"github.com/fdanis/ygtrack/cmd/server/store/repository"
	"github.com/go-chi/chi"
)

type MetricHandler struct {
	CounterRepo repository.MetricRepository[int64]
	GaugeRepo   repository.MetricRepository[float64]
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")
	valueMetric := chi.URLParam(r, "value")

	switch typeMetric {
	case "gauge":
		val, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.GaugeRepo.Add(dataclass.Metric[float64]{Name: nameMetric, Value: val})
	case "counter":
		val, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.CounterRepo.Add(dataclass.Metric[int64]{Name: nameMetric, Value: val})
	default:
		http.Error(w, "Incorrect type", http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")

	result := ""
	switch typeMetric {
	case "gauge":
		met, err := h.GaugeRepo.GetByName(nameMetric)
		if err != nil {
			log.Print(err)
		}
		if met == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		result = fmt.Sprintf("%.3f", met.Value)

	case "counter":
		met, err := h.CounterRepo.GetByName(nameMetric)
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

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {

	counterList, err := h.CounterRepo.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	gaugeList, err := h.GaugeRepo.GetAll()
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
	render.Render(w, "home.html", &models.TemplateDate{Data: map[string]any{"metrics": result}})
	w.WriteHeader(http.StatusOK)
}
