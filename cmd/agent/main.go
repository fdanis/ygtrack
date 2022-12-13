package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fdanis/ygtrack/cmd/agent/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

const (
	PollInterval   int    = 2
	ReportInterval int    = 10
	ServerURL      string = "localhost:8080"
)

type Conf struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"12"`
}

func NewConfig() *Conf {
	config := Conf{}
	//почемуто тестировщики в яндексе передают в REPORT_INTERVAL вместо '2' корявую строку '2s' и это не работает
	//2022/12/14 00:12:32 env: parse error on field "PollInterval" of type "int64": strconv.ParseInt: parsing "2s": invalid syntax; parse error on field "ReportInterval" of type "int64": strconv.ParseInt: parsing "10s": invalid syntax
	err := env.Parse(&config)
	if err != nil {
		log.Println(err)
	}
	// а теперь получаем конфиги стандартным способом
	config.Address = os.Getenv("ADDRESS")
	if config.Address == "" {
		config.Address = ServerURL
	}

	if p, err := strconv.ParseInt(os.Getenv("POLL_INTERVAL"), 10, 64); err == nil && p > 0 {
		config.PollInterval = p
	} else {
		config.PollInterval = int64(PollInterval)
	}

	if p, err := strconv.ParseInt(os.Getenv("REPORT_INTERVAL"), 10, 64); err == nil && p > 0 {
		config.ReportInterval = p
	} else {
		config.ReportInterval = int64(ReportInterval)
	}
	return &config
}

func main() {
	config := NewConfig()

	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)

	ctxupdate, cancelu := context.WithCancel(context.Background())
	ctxsend, cancels := context.WithCancel(context.Background())
	go Update(ctxupdate, config.PollInterval, m)
	go Send(ctxsend, config.ReportInterval, config.Address, m)

	defer cancelu()
	defer cancels()
	for {
		if false {
			break
		}
	}
}
func Exit(cancel context.CancelFunc) {
	bufio.NewReader(os.Stdin).ReadBytes('q')
	cancel()
}

func Update(ctx context.Context, poolInterval int64, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(poolInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Update()
		case <-ctx.Done():
			{
				//why I can't see this line in console?
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
func Send(ctx context.Context, sendInterval int64, host string, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(sendInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Send("http://" + strings.TrimRight(host, "/") + "/update")
		case <-ctx.Done():
			{
				//why I can't see this line in console?
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
