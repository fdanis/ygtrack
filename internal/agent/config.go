package agent

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env"
)

type Conf struct {
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Key            string        `env:"KEY"`
}

func ReadEnv(config *Conf) {
	err := env.Parse(config)
	if err != nil {
		log.Println(err)
	}
}

func ReadFlags(config *Conf) {
	flag.StringVar(&config.Address, "a", "localhost:8080", "host for server")
	flag.DurationVar(&config.PollInterval, "p", time.Second*2, "interval fo pooling metrics")
	flag.DurationVar(&config.ReportInterval, "r", time.Second*10, "interval fo report")
	flag.StringVar(&config.Key, "k", "", "key for hash function")
}
