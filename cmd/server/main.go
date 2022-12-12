package main

import (
	"fmt"
	"log"
	"os"

	//"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/handler"
	"github.com/fdanis/ygtrack/internal/server/render"

	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/go-chi/chi"
)

var app config.AppConfig

func main() {

	// cachecdTemplate, err := render.CreateTemplateCache()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// app.TemplateCache = cachecdTemplate
	app.UseTemplateCache = false
	render.NewTemplates(&app)
	cr := metricrepository.NewMetricRepository[int64]()
	gr := metricrepository.NewMetricRepository[float64]()
	metricHandler := handler.MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", metricHandler.Update)
	r.Post("/update/", metricHandler.UpdateJSON)
	r.Post("/update", metricHandler.UpdateJSON)
	r.Post("/value/", metricHandler.GetJSONValue)
	r.Post("/value", metricHandler.GetJSONValue)
	r.Get("/value/{type}/{name}", metricHandler.GetValue)
	r.Get("/", metricHandler.Get)

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "localhost:8080"
	}
	server := &http.Server{
		Addr:    address,
		Handler: r,
	}
	fmt.Println(address)
	log.Println(address)
	server.ListenAndServe()
}
