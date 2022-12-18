package config

import (
	"flag"
	"html/template"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
)

type AppConfig struct {
	UseTemplateCache  bool
	TemplateCache     map[string]*template.Template
	EnvConfig         EnvConfig
	CounterRepository metricrepository.MetricRepository[int64]
	GaugeRepository   metricrepository.MetricRepository[float64]
	ChForSyncWithFile chan int
	SaveToFileSync    bool
}

type EnvConfig struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func (c *EnvConfig) ReadEnv() error {
	err := env.Parse(c)
	if err != nil {
		return err
	}
	return nil
}

func (c *EnvConfig)ReadFlags() {
	flag.StringVar(&c.Address, "a", ":8080", "host for server")
	flag.BoolVar(&c.Restore, "r", false, "restore data from file")
	flag.DurationVar(&c.StoreInterval, "i", time.Second*2, "interval fo saving data to file")
	flag.StringVar(&c.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path")
}