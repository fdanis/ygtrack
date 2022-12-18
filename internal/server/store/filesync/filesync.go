package filesync

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

func FileSync(app *config.AppConfig, ctx context.Context) {
	if app.EnvConfig.StoreFile != "" {
		os.Mkdir(path.Dir(app.EnvConfig.StoreFile), 0777)
		if app.EnvConfig.Restore {
			loadFromFile(app.EnvConfig.StoreFile, &app.GaugeRepository, &app.CounterRepository)
		}
		if app.EnvConfig.StoreInterval != 0 {
			go syncByInterval(app.ChForSyncWithFile, ctx, app.EnvConfig.StoreInterval)
		} else {
			app.SaveToFileSync = true
		}
		go sync(app.EnvConfig.StoreFile, app.ChForSyncWithFile, ctx, &app.CounterRepository, &app.GaugeRepository)
	}
}

func syncByInterval(ch chan int, ctx context.Context, interavl time.Duration) {
	t := time.NewTicker(interavl)
	for {
		select {
		case <-t.C:
			ch <- 1
		case <-ctx.Done():
			{
				t.Stop()
				return
			}
		}
	}
}

func sync(fileName string, ch chan int, ctx context.Context, counterRepo repository.MetricRepository[int64], gaugeRepo repository.MetricRepository[float64]) {
	for {
		select {
		case <-ch:
			writeToFile(fileName, gaugeRepo, counterRepo)
		case <-ctx.Done():
			{
				writeToFile(fileName, gaugeRepo, counterRepo)
				return
			}
		}
	}
}

func writeToFile(fileName string, gaugeRepo repository.MetricRepository[float64], counterRepo repository.MetricRepository[int64]) {

	file, err := os.OpenFile(fileName, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0777)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	g, err := gaugeRepo.GetAll()
	if err != nil {
		log.Println(err)
	}
	enc.Encode(g)

	c, err := counterRepo.GetAll()
	if err != nil {
		log.Println(err)
	}
	enc.Encode(c)
}

func loadFromFile(fileName string, gaugeRepo repository.MetricRepository[float64], counterRepo repository.MetricRepository[int64]) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	enc := json.NewDecoder(file)
	var val []dataclass.Metric[float64]
	enc.Decode(&val)
	for _, item := range val {
		gaugeRepo.Add(item)
	}
	var cval []dataclass.Metric[int64]
	enc.Decode(&cval)
	for _, item := range cval {
		counterRepo.Add(item)
	}
}
