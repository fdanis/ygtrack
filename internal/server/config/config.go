package config

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fdanis/ygtrack/internal/server/store/filesync"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
	_ "github.com/mattn/go-sqlite3"
)

type AppConfig struct {
	UseTemplateCache  bool
	TemplateCache     map[string]*template.Template
	Parameters        Environment
	CounterRepository repository.MetricRepository[int64]
	GaugeRepository   repository.MetricRepository[float64]
	ChForSyncWithFile chan int
	SaveToFileSync    bool
}

type Environment struct {
	Address          string        `env:"ADDRESS" json:"address"`
	StoreInterval    time.Duration `env:"STORE_INTERVAL" json:"store_interval"`
	StoreFile        string        `env:"STORE_FILE" json:"store_file"`
	Restore          bool          `env:"RESTORE" json:"restore"`
	Key              string        `env:"KEY"`
	ConnectionString string        `env:"DATABASE_DSN" json:"database_dsn"`
	CryptoKey        string        `env:"CRYPTO_KEY" json:"crypto_key"`
}

func (c *Environment) ReadEnv() error {
	return env.Parse(c)
}

func (c *Environment) ReadFlags() string {

	flag.StringVar(&c.Address, "a", ":8080", "host for server")
	flag.BoolVar(&c.Restore, "r", false, "restore data from file")
	flag.DurationVar(&c.StoreInterval, "i", time.Second*2, "interval fo saving data to file")
	flag.StringVar(&c.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path")
	flag.StringVar(&c.Key, "k", "", "hash key")
	flag.StringVar(&c.ConnectionString, "d", "", "connection string")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "crypto-key")
	file := ""
	flag.StringVar(&file, "c", "", "file for config")
	return file
}

func (app *AppConfig) InitFileStorage(ctx context.Context) error {
	if app.Parameters.StoreFile != "" {
		if _, err := os.Stat(path.Dir(app.Parameters.StoreFile)); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path.Dir(app.Parameters.StoreFile), 0777)
			if err != nil {
				fmt.Println("create dirrrectory")
				return err
			}
		}

		if app.Parameters.Restore {
			err := filesync.LoadFromFile(app.Parameters.StoreFile, app.GaugeRepository, app.CounterRepository)
			if err != nil {
				fmt.Println("load from file")
				return err
			}
		}
		if app.Parameters.StoreInterval != 0 {
			go filesync.SyncByInterval(ctx, app.ChForSyncWithFile, app.Parameters.StoreInterval)
		} else {
			app.SaveToFileSync = true
		}
		go filesync.Sync(ctx, app.Parameters.StoreFile, app.ChForSyncWithFile, app.CounterRepository, app.GaugeRepository)
	}
	return nil
}

func (app *AppConfig) InitDataBaseStorage() error {

	db, err := sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Println(err)
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.counter (created timestamp PRIMARY KEY, val integer);")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.counter (created timestamp, name varchar(50), val integer);")
	if err != nil {
		return err
	}

	if app.Parameters.StoreFile != "" {
		if _, err := os.Stat(path.Dir(app.Parameters.StoreFile)); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path.Dir(app.Parameters.StoreFile), 0777)
			if err != nil {
				fmt.Println("create dirrrectory")
				return err
			}
		}

		if app.Parameters.Restore {
			err := filesync.LoadFromFile(app.Parameters.StoreFile, app.GaugeRepository, app.CounterRepository)
			if err != nil {
				fmt.Println("load from file")
				return err
			}
		}
		if app.Parameters.StoreInterval != 0 {
			go filesync.SyncByInterval(ctx, app.ChForSyncWithFile, app.Parameters.StoreInterval)
		} else {
			app.SaveToFileSync = true
		}
		go filesync.Sync(ctx, app.Parameters.StoreFile, app.ChForSyncWithFile, app.CounterRepository, app.GaugeRepository)
	}
	return nil
}

func (c *Environment) LoadFromConfigFile(file string) {
	if file != "" {
		var tmpConf Environment
		data, err := os.ReadFile(file)
		if err != nil {
			panic("config file does not exists")
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		if err = dec.Decode(&tmpConf); err != nil {
			panic("config file not correct")
		}
		if c.Address == "" {
			c.Address = tmpConf.Address
		}
		if c.CryptoKey == "" {
			c.CryptoKey = tmpConf.CryptoKey
		}
		if c.StoreInterval == 0 {
			c.StoreInterval = tmpConf.StoreInterval
		}
		if c.ConnectionString == "" {
			c.ConnectionString = tmpConf.ConnectionString
		}
		if c.StoreFile == "" {
			c.StoreFile = tmpConf.StoreFile
		}
		if c.CryptoKey == "" {
			c.StoreFile = tmpConf.CryptoKey
		}
	}
}
