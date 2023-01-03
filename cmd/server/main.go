package main

import (
	"context"

	"flag"
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/driver"
	"github.com/fdanis/ygtrack/internal/server"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/fdanis/ygtrack/internal/server/store/repository/pgxmetricrepository"
)

func main() {
	app := config.AppConfig{}
	//read environments
	app.Parameters.ReadFlags()
	flag.Parse()
	err := app.Parameters.ReadEnv()
	if err != nil {
		log.Println("Read Env Error")
		log.Fatalln(err)
	}

	var db *driver.DB
	app.Parameters.ConnectionString = "postgres://RxAdviceServices:cbltkfDjhjyf1@localhost:5432/temp"
	if app.Parameters.ConnectionString != "" {
		db, err = driver.ConnectSQL(app.Parameters.ConnectionString)
		if err != nil {
			panic(err)
		}
	}

	//initialize html template
	cachecdTemplate, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalln(err)
	}
	app.TemplateCache = cachecdTemplate
	app.UseTemplateCache = true
	render.NewTemplates(&app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if db != nil {
		app.CounterRepository = pgxmetricrepository.NewCountRepository(db.SQL)
		app.GaugeRepository = pgxmetricrepository.NewGougeRepository(db.SQL)
	} else {
		app.CounterRepository = metricrepository.NewMetricRepository[int64]()
		app.GaugeRepository = metricrepository.NewMetricRepository[float64]()
		app.ChForSyncWithFile = make(chan int)

		//synchronization with file
		err = app.InitFileStorage(ctx)
		if err != nil {
			log.Println("FileSync Error")
			log.Println(err)
		}
	}
	server := &http.Server{
		Addr:    app.Parameters.Address,
		Handler: server.Routes(&app),
	}
	log.Println("server started")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	if db != nil {
		db.SQL.Close()
	}
	log.Println("server stoped")
}
