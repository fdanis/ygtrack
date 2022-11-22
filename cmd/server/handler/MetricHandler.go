package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
	"github.com/fdanis/ygtrack/cmd/server/store/repository"
)

type MetricHandler struct {
	CounterRepo repository.CounterRepository
	GaugeRepo   repository.GaugeRepository
}

func (h MetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not found", http.StatusNotFound)
		return
	}

	url := strings.Trim(r.URL.Path, "/")
	urlitem := strings.Split(url, "/")
	if len(urlitem) != 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch urlitem[1] {
	case "gauge":
		val, err := strconv.ParseFloat(urlitem[3], 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.GaugeRepo.Add(dataclass.GaugeMetric{Name: urlitem[2], Value: val})
		break
	case "counter":
		val, err := strconv.ParseUint(urlitem[3], 10, 64)
		if err != nil {
			http.Error(w, "Incorrect value", http.StatusBadRequest)
			return
		}
		h.CounterRepo.Add(dataclass.CounterMetric{Name: urlitem[2], Value: val})
		break
	default:
		http.Error(w, "Incorrect type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
