package config

import (
	"context"
	"flag"
	"html/template"
	"os"
	"path"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fdanis/ygtrack/internal/server/store/filesync"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
)

type AppConfig struct {
	UseTemplateCache  bool
	TemplateCache     map[string]*template.Template
	EnvConfig         Environment
	CounterRepository metricrepository.MetricRepository[int64]
	GaugeRepository   metricrepository.MetricRepository[float64]
	ChForSyncWithFile chan int
	SaveToFileSync    bool
}

type Environment struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func (c *Environment) ReadEnv() error {
	err := env.Parse(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Environment) ReadFlags() {
	flag.StringVar(&c.Address, "a", ":8080", "host for server")
	flag.BoolVar(&c.Restore, "r", false, "restore data from file")
	flag.DurationVar(&c.StoreInterval, "i", time.Second*2, "interval fo saving data to file")
	flag.StringVar(&c.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path")
}

func (app *AppConfig) FileSync(ctx context.Context) error {
	if app.EnvConfig.StoreFile != "" {
		err := os.Mkdir(path.Dir(app.EnvConfig.StoreFile), 0777)
		if err != nil {
			return err
		}
		if app.EnvConfig.Restore {
			err = filesync.LoadFromFile(app.EnvConfig.StoreFile, &app.GaugeRepository, &app.CounterRepository)
			if err != nil {
				return err
			}
		}
		if app.EnvConfig.StoreInterval != 0 {
			go filesync.SyncByInterval(ctx, app.ChForSyncWithFile, app.EnvConfig.StoreInterval)
		} else {
			app.SaveToFileSync = true
		}
		go filesync.Sync(ctx, app.EnvConfig.StoreFile, app.ChForSyncWithFile, &app.CounterRepository, &app.GaugeRepository)
	}
	return nil
}
