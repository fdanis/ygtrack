package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/server"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
)

func main() {
	app := config.AppConfig{}
	//read environments
	app.EnvConfig.ReadFlags()
	flag.Parse()
	app.EnvConfig.ReadEnv()

	//initialize html template
	cachecdTemplate, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalln(err)
	}
	app.TemplateCache = cachecdTemplate
	app.UseTemplateCache = true
	app.CounterRepository = metricrepository.NewMetricRepository[int64]()
	app.GaugeRepository = metricrepository.NewMetricRepository[float64]()
	app.ChForSyncWithFile = make(chan int)
	render.NewTemplates(&app)

	//synchronization with file
	ctx, cancel := context.WithCancel(context.Background())
	app.FileSync(ctx)
	defer cancel()

	server := &http.Server{
		Addr:    app.EnvConfig.Address,
		Handler: server.Routes(&app),
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
