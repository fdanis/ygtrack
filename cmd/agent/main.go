package main

import (
	"bufio"
	"context"
	"flag"
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
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func readEnv(config *Conf) {
	err := env.Parse(config)
	if err != nil {
		log.Println(err)
	}
}

func readFlags(config *Conf) {
	flag.StringVar(&config.Address, "a", "localhost:8080", "host for server")
	flag.DurationVar(&config.PollInterval, "p", time.Second*2, "interval fo pooling metrics")
	flag.DurationVar(&config.ReportInterval, "r", time.Second*10, "interval fo report")
}

func main() {
	config := Conf{}
	readFlags(&config)
	flag.Parse()
	readEnv(&config)
	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)
	ctx, cancel := context.WithCancel(context.Background())
	go Update(ctx, config.PollInterval, m)
	go Send(ctx, config.ReportInterval, config.Address, m)
	defer cancel()
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
