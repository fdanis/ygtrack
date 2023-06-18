package agent

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
)

type Conf struct {
	Address        string        `env:"ADDRESS" json:"address"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	CryptoKey      string        `env:"CRYPTO_KEY" json:"crypto_key"`
	Key            string        `env:"KEY"`
	UseGrpc        bool          `env:"USE_GRPC"`
}

func ReadEnv(config *Conf) error {
	err := env.Parse(config)
	if err != nil {
		return err
	}
	return nil
}

// ReadFlags inits flags and return config path
func ReadFlags(config *Conf) *string {
	flag.StringVar(&config.Address, "a", "localhost:8080", "host for server")
	flag.StringVar(&config.Key, "k", "", "key for hash function")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "key for hash function")
	flag.DurationVar(&config.PollInterval, "p", time.Second*2, "interval fo pooling metrics")
	flag.DurationVar(&config.ReportInterval, "r", time.Second*10, "interval fo report")
	flag.BoolVar(&config.UseGrpc, "g", false, "use grpc connection")

	file := ""
	flag.StringVar(&file, "c", "", "file for config")
	return &file
}

func (c *Conf) LoadFromConfigFile(file string) {
	if file != "" {
		var tmpConf Conf
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
		if c.PollInterval == 0 {
			c.PollInterval = tmpConf.PollInterval
		}
		if c.ReportInterval == 0 {
			c.ReportInterval = tmpConf.ReportInterval
		}
	}
}

func (c *Conf) UnmarshalJSON(data []byte) error {
	type ConfAlias Conf
	aliasValue := &struct {
		*ConfAlias
		PollInterval   string `json:"poll_interval"`
		ReportInterval string `json:"report_interval"`
	}{
		ConfAlias: (*ConfAlias)(c),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}
	p, err := time.ParseDuration(aliasValue.PollInterval)
	if err != nil {
		return err
	}
	c.PollInterval = p
	r, err := time.ParseDuration(aliasValue.ReportInterval)
	if err != nil {
		return err
	}
	c.PollInterval = r
	return nil
}
