package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/handler"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
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
	env, err := config.NewEnvConfig()
	if err != nil {
		log.Fatal(err)
	}
	app.EnvConfig = env
	render.NewTemplates(&app)

	cr := metricrepository.NewMetricRepository[int64]()
	gr := metricrepository.NewMetricRepository[float64]()

	metricHandler := handler.NewMetricHandler(&cr, &gr)
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", metricHandler.Update)
	r.Post("/update/", metricHandler.UpdateJSON)
	r.Post("/update", metricHandler.UpdateJSON)
	r.Post("/value/", metricHandler.GetJSONValue)
	r.Post("/value", metricHandler.GetJSONValue)
	r.Get("/value/{type}/{name}", metricHandler.GetValue)
	r.Get("/", metricHandler.Get)

	ctx, cancel := context.WithCancel(context.Background())
	if app.EnvConfig.StoreFile != "" && app.EnvConfig.StoreInterval != 0 {
		os.Mkdir(path.Dir(app.EnvConfig.StoreFile), 0777)
		go Sync(ctx, app, &cr, &gr)
	}
	defer cancel()
	server := &http.Server{
		Addr:    app.EnvConfig.Address,
		Handler: r,
	}
	server.ListenAndServe()
}

func Sync(ctx context.Context, app config.AppConfig, counterRepo repository.MetricRepository[int64], gaugeRepo repository.MetricRepository[float64]) {
	t := time.NewTicker(app.EnvConfig.StoreInterval)
	for {
		select {
		case <-t.C:
			file, err := os.OpenFile(app.EnvConfig.StoreFile, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0777)

			fmt.Println(file.Name())
			if err != nil {
				log.Println(err)
			}
			defer file.Close()
			enc := json.NewEncoder(file)
			g, err := gaugeRepo.GetAll()
			if err != nil {
				log.Println(err)
			}
			for _, item := range g {
				enc.Encode(models.Metrics{ID: item.Name, MType: "gauge", Value: &item.Value})
			}
			c, err := counterRepo.GetAll()
			if err != nil {
				log.Println(err)
			}
			for _, item := range c {
				enc.Encode(models.Metrics{ID: item.Name, MType: "counter", Delta: &item.Value})
			}

		case <-ctx.Done():
			{
				t.Stop()
				return
			}
		}
	}
}
