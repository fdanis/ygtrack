package config

import (
	"html/template"
	"time"

	"github.com/caarlos0/env/v6"
)

type AppConfig struct {
	UseTemplateCache bool
	TemplateCache    map[string]*template.Template
	EnvConfig        EnvConfig
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
