package main

import (
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/cmd/server/config"
	"github.com/fdanis/ygtrack/cmd/server/handler"
	"github.com/fdanis/ygtrack/cmd/server/render"
	"github.com/fdanis/ygtrack/cmd/server/store/repository/metricrepository"
	"github.com/go-chi/chi"
)

var app config.AppConfig

func main() {

	cachecdTemplate, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalln(err)
	}

	app.TemplateCache = cachecdTemplate
	app.UseTemplateCache = true

	render.NewTemplates(&app)
	cr := metricrepository.NewMetricRepository[int64]()
	gr := metricrepository.NewMetricRepository[float64]()
	metricHandler := handler.MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", metricHandler.Update)
	r.Get("/value/{type}/{name}", metricHandler.GetValue)
	r.Get("/", metricHandler.Get)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	server.ListenAndServe()

}
