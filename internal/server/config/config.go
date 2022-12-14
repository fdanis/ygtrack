package config

import (
	"html/template"
	"time"

	"github.com/caarlos0/env/v6"
)

type AppConfig struct {
	UseTemplateCache bool
	TemplateCache    map[string]*template.Template
	EnvConfig        *EnvConfig
}

type EnvConfig struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"15s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

func NewEnvConfig() (*EnvConfig, error) {
	config := EnvConfig{}
	err := env.Parse(&config)
	if err != nil {
		return &config, err
	}
	return &config, nil
}
