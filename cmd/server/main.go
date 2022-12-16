package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/handler"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/filesync"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/go-chi/chi"
)

var app config.AppConfig

func readFlags(c *config.EnvConfig) {
	flag.StringVar(&c.Address, "a", ":8080", "host for server")
	flag.BoolVar(&c.Restore, "r", false, "restore data from file")
	flag.DurationVar(&c.StoreInterval, "i", time.Second*2, "interval fo saving data to file")
	flag.StringVar(&c.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path")
}

func main() {
	//read environments
	readFlags(&app.EnvConfig)
	flag.Parse()
	fmt.Printf("%v", app.EnvConfig)
	err := app.EnvConfig.ReadEnv()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", app.EnvConfig)

	//initialize html template
	cachecdTemplate, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalln(err)
	}
	app.TemplateCache = cachecdTemplate
	app.UseTemplateCache = true
	render.NewTemplates(&app)

	//initialize handler
	cr := metricrepository.NewMetricRepository[int64]()
	gr := metricrepository.NewMetricRepository[float64]()
	ch := make(chan int)
	metricHandler := handler.NewMetricHandler(&cr, &gr)
	//route
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", metricHandler.Update)
	r.Post("/update/", metricHandler.UpdateJSON)
	r.Post("/update", metricHandler.UpdateJSON)
	r.Post("/value/", metricHandler.GetJSONValue)
	r.Post("/value", metricHandler.GetJSONValue)
	r.Get("/value/{type}/{name}", metricHandler.GetValue)
	r.Get("/", metricHandler.Get)

	//synchronization with file
	ctx, cancel := context.WithCancel(context.Background())
	ctxSync, cancelSync := context.WithCancel(context.Background())
	if app.EnvConfig.StoreFile != "" {
		os.Mkdir(path.Dir(app.EnvConfig.StoreFile), 0777)
		if app.EnvConfig.Restore {
			filesync.LoadFromFile(app.EnvConfig.StoreFile, &gr, &cr)
		}
		if app.EnvConfig.StoreInterval != 0 {
			go filesync.SyncByInterval(ch, ctx, app.EnvConfig.StoreInterval)
		} else {
			metricHandler.Ch = ch
		}
		go filesync.Sync(app.EnvConfig.StoreFile, ch, ctxSync, &cr, &gr)
	}

	defer cancel()
	defer cancelSync()
	server := &http.Server{
		Addr:    app.EnvConfig.Address,
		Handler: r,
	}
	server.ListenAndServe()
}
