package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
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
		oldValue, err := h.CounterRepo.GetByName(nameMetric)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusInternalServerError)
			return
		}
		if oldValue != nil {
			val += oldValue.Value
		}
		h.CounterRepo.Add(dataclass.Metric[int64]{Name: nameMetric, Value: val})
	default:
		http.Error(w, "Incorrect type", http.StatusNotImplemented)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	var model models.Metrics
	if err := decodeJSONBody(w, r, &model); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	switch model.MType {
	case "counter":
		h.CounterRepo.Add(dataclass.Metric[int64]{Name: model.ID, Value: *model.Delta})
	case "gauge":
		h.GaugeRepo.Add(dataclass.Metric[float64]{Name: model.ID, Value: *model.Value})
	}
	responseJSON(w, &model)
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

func (h *MetricHandler) GetJSONValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var model models.Metrics
	if err := decodeJSONBody(w, r, &model); err != nil {
		var mr *malformedRequest
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
		met, err := h.GaugeRepo.GetByName(model.ID)
		if err != nil {
			log.Print(err)
		}
		if met != nil {
			model.Value = &met.Value
		}

	case "counter":
		met, err := h.CounterRepo.GetByName(model.ID)
		if err != nil {
			log.Print(err)
		}
		if met != nil {
			model.Delta = &met.Value
		}

	}

	responseJSON(w, model)
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

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func responseJSON(w http.ResponseWriter, src interface{}) {

	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(src)
	w.Write(buf.Bytes())
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		if strings.ToLower(contentType) != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
