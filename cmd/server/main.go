package main

import (
	"net/http"

	"github.com/fdanis/ygtrack/cmd/server/handler"
	"github.com/fdanis/ygtrack/cmd/server/store/repository/counterrepository"
	"github.com/fdanis/ygtrack/cmd/server/store/repository/gaugerepository"
)

func main() {
	cr := counterrepository.NewCounterRepository()
	gr := gaugerepository.NewGaugeRepository()
	metricHandler := handler.MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}

	http.Handle("/update/", metricHandler)

	server := &http.Server{
		Addr: ":8080",
	}
	server.ListenAndServe()
}
