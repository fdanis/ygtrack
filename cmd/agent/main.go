package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/fdanis/ygtrack/cmd/agent/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

const (
	PollInterval   int    = 2
	ReportInterval int    = 10
	ServerURL      string = "127.0.0.1:8080"
)

type conf struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	var config = conf{}
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)

	//ctxupdate, cancelu := context.WithCancel(context.Background())
	//ctxsend, cancels := context.WithCancel(context.Background())
	go Update(config.PollInterval, m)
	go Send(config.ReportInterval, os.Getenv("ADDRESS"), m)
	for {
		if false {
			break
		}
	}
	//defer cancelu()
	//defer cancels()
}
func Exit(cancel context.CancelFunc) {
	bufio.NewReader(os.Stdin).ReadBytes('q')
	cancel()
}

func Update(poolInterval int, service *memstatservice.SimpleMemStatService) {
	if poolInterval <= 0 {
		poolInterval = PollInterval
	}

	t := time.NewTicker(time.Duration(poolInterval) * time.Second)
	
	for {
		select {
		case <-t.C:
			service.Update()
		}
	}
}
func Send(sendInterval int, host string, service *memstatservice.SimpleMemStatService) {
	if sendInterval <= 0 {
		sendInterval = ReportInterval
	}

	if host == "" {
		host = ServerURL
	}
	t := time.NewTicker(time.Duration(sendInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Send("http://" + strings.TrimRight(host, "/") + "/update")
		}
	}
}
