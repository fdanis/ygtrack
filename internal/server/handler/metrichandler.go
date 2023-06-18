package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/helpers"
	ms "github.com/fdanis/ygtrack/internal/server/metricsservice"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/go-chi/chi"
)

type DbChecker interface {
	Ping() error
}

// MetricHandler - structure for handling metrics
type MetricHandler struct {
	service   *ms.MetricsService
	hashkey   string
	dbchecker DbChecker
}

func NewMetricHandler(service *ms.MetricsService, hashkey string, dbchecker DbChecker) MetricHandler {
	result := MetricHandler{
		service:   service,
		hashkey:   hashkey,
		dbchecker: dbchecker,
	}
	return result
}

// Update - GET request for updating metrics from queryParams
func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	var model models.Metrics

	model.MType = strings.ToLower(chi.URLParam(r, "type"))
	model.ID = chi.URLParam(r, "name")
	valueMetric := chi.URLParam(r, "value")
	switch model.MType {
	case constants.MetricsTypeGauge:

		val, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		model.Value = &val
	case constants.MetricsTypeCounter:
		val, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		model.Delta = &val
	default:
		http.Error(w, "Incorrect type", http.StatusNotImplemented)
		return
	}

	err := h.service.AddMetric(model)
	if err != nil {
		var merr *ms.MetricsError
		if errors.As(err, &merr) {
			http.Error(w, merr.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateJSON - POST method for updating metrics by json
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

	err := h.service.AddMetric(model)
	if err != nil {
		var merr *ms.MetricsError
		if errors.As(err, &merr) {
			http.Error(w, merr.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}
	responseJSON(w, &model)
}

// UpdateBatch - POST request. Update batch metrics
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

	err := h.service.UpdateBatch(model)
	if err != nil {
		var merr *ms.MetricsError
		if errors.As(err, &merr) {
			http.Error(w, merr.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	typeMetric := strings.ToLower(chi.URLParam(r, "type"))
	nameMetric := chi.URLParam(r, "name")
	result := ""
	switch typeMetric {
	case constants.MetricsTypeGauge:
		met, err := h.service.GetGaugeValue(nameMetric)
		if err != nil {
			var merr *ms.MetricsError
			if errors.As(err, &merr) {
				if merr.Code == 1 {
					w.WriteHeader(http.StatusNotFound)
				} else {
					log.Print(err)
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
		}
		result = fmt.Sprintf("%.3f", met)
	case constants.MetricsTypeCounter:
		met, err := h.service.GetCounterValue(nameMetric)
		if err != nil {
			var merr *ms.MetricsError
			if errors.As(err, &merr) {
				if merr.Code == 1 {
					w.WriteHeader(http.StatusNotFound)
				} else {
					log.Print(err)
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
		}
		result = fmt.Sprintf("%d", met)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// GetJSONValue - get metrics json format
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
		met, err := h.service.GetGaugeValue(model.ID)
		if err != nil {
			var merr *ms.MetricsError
			if errors.As(err, &merr) {
				if merr.Code == 1 {
					w.WriteHeader(http.StatusNotFound)
				} else {
					log.Print(err)
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
		}
		model.Value = &met
	case constants.MetricsTypeCounter:
		met, err := h.service.GetCounterValue(model.ID)
		if err != nil {
			var merr *ms.MetricsError
			if errors.As(err, &merr) {
				if merr.Code == 1 {
					w.WriteHeader(http.StatusNotFound)
				} else {
					log.Print(err)
					http.Error(w, "Server error", http.StatusInternalServerError)
				}
				return
			}
		}
		model.Delta = &met
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	hash, err := helpers.GetHash(model, h.hashkey)
	if err != nil {
		log.Println(err)
	}
	model.Hash = hash
	responseJSON(w, model)
}

// Get - get request for getting all metrics in html
func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	counterList, err := h.service.GetAllCounter()
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
	gaugeList, err := h.service.GetAllGauge()
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error", http.StatusInternalServerError)
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

// Ping - request for ping DB
func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.dbchecker.Ping()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
