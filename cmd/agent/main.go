package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fdanis/ygtrack/cmd/agent/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

type Conf struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
}

func NewConfig() *Conf {
	config := Conf{}
	err := env.Parse(&config)
	if err != nil {
		log.Println(err)
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

func Update(ctx context.Context, poolInterval time.Duration, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(poolInterval)
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
func Send(ctx context.Context, sendInterval time.Duration, host string, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(sendInterval)
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
