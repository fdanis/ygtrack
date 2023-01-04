package filesync

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

func SyncByInterval(ctx context.Context, ch chan int, interavl time.Duration) {
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

func Sync(ctx context.Context, fileName string, ch chan int, counterRepo repository.MetricRepository[int64], gaugeRepo repository.MetricRepository[float64]) {
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

func LoadFromFile(fileName string, gaugeRepo repository.MetricRepository[float64], counterRepo repository.MetricRepository[int64]) error {
	if _, err := os.Stat(fileName); err != nil {
		return nil
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewDecoder(file)
	var val []dataclass.Metric[float64]
	err = enc.Decode(&val)
	if err != nil {
		return err
	}
	for _, item := range val {
		gaugeRepo.Add(item)
	}
	var cval []dataclass.Metric[int64]
	enc.Decode(&cval)
	for _, item := range cval {
		counterRepo.Add(item)
	}
	return nil
}
