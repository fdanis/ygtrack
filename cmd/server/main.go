package main

import (
	"context"
	"database/sql"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/driver"
	"github.com/fdanis/ygtrack/internal/server"
	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/grpcservice"
	"github.com/fdanis/ygtrack/internal/server/metricsservice"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/fdanis/ygtrack/internal/server/store/repository/pgxmetricrepository"
	"google.golang.org/grpc"

	_ "net/http/pprof"

	pb "github.com/fdanis/ygtrack/proto"

	_ "github.com/golang-migrate/migrate/source/file"
)

//go:generate go run ../generator/genvar.go string
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printInfoVar()
	app := config.AppConfig{}
	//read environments
	file := app.Parameters.ReadFlags()
	flag.Parse()
	err := app.Parameters.ReadEnv()
	if err != nil {
		log.Println("Read Env Error")
		log.Fatalln(err)
	}
	app.Parameters.LoadFromConfigFile(*file)

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
	metricservice := metricsservice.NewMetricsService(&app, db)
	server := &http.Server{
		Addr:    app.Parameters.Address,
		Handler: server.Routes(&app, db, metricservice),
	}

	log.Printf("server started at %s\n", app.Parameters.Address)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", app.Parameters.GrpcAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterMetricServiceServer(grpcServer, grpcservice.NewGrpcMetricServer(metricservice))
	go func() {
		log.Printf("grpc server started at %s\n", app.Parameters.GrpcAddress)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sig
	server.Shutdown(ctx)
	// wait 2 second for server shutdown
	time.Sleep(time.Second * 2)
}
