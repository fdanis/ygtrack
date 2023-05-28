package main

import (
	"context"
	"database/sql"

	"flag"
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/driver"
	"github.com/fdanis/ygtrack/internal/server"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/fdanis/ygtrack/internal/server/store/repository/pgxmetricrepository"

	_ "net/http/pprof"

	_ "github.com/golang-migrate/migrate/source/file"
)

//go:generate go run ../generator/genvar.go string
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printInfoVar()
	app := config.AppConfig{}
	//read environments
	app.Parameters.ReadFlags()
	flag.Parse()
	err := app.Parameters.ReadEnv()
	if err != nil {
		log.Println("Read Env Error")
		log.Fatalln(err)
	}

	var db *sql.DB
	if app.Parameters.ConnectionString != "" {
		db, err = driver.ConnectSQL(app.Parameters.ConnectionString)
		if err != nil {
			panic(err)
		}
		defer db.Close()
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
		app.CounterRepository = pgxmetricrepository.NewCountRepository(db)
		app.GaugeRepository = pgxmetricrepository.NewGougeRepository(db)
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
		Handler: server.Routes(&app, db),
	}

	log.Printf("server started at %s\n", app.Parameters.Address)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
