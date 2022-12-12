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
	Address        string `env:"ADDRESS"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
}

func main() {
	config := Conf{}
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	config.Address = os.Getenv("ADDRESS")
	config.PollInterval, err = strconv.ParseInt(os.Getenv("POLL_INTERVAL"), 10, 64)
	config.ReportInterval, err = strconv.ParseInt(os.Getenv("REPORT_INTERVAL"), 10, 64)

	fmt.Printf("%v", config)
	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)

	//ctxupdate, cancelu := context.WithCancel(context.Background())
	//ctxsend, cancels := context.WithCancel(context.Background())
	go Update(config.PollInterval, m)
	go Send(config.ReportInterval, config.Address, m)
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

func Update(poolInterval int64, service *memstatservice.SimpleMemStatService) {

	t := time.NewTicker(time.Duration(poolInterval) * time.Second)

	for {
		select {
		case <-t.C:
			service.Update()
		}
	}
}
func Send(sendInterval int64, host string, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(sendInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Send("http://" + strings.TrimRight(host, "/") + "/update")
		}
	}
}
