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

type conf struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

var config = conf{Address: "http://localhost:8080", ReportInterval: 10, PollInterval: 2}

func main() {
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)
	ctxupdate, cancelu := context.WithCancel(context.Background())
	ctxsend, cancels := context.WithCancel(context.Background())
	//	ctxend, cancele := context.WithCancel(context.Background())
	go Update(ctxupdate, config.PollInterval, m)
	go Send(ctxsend, config.ReportInterval, strings.TrimRight(config.Address, "/")+"/update", m)

	for {
		time.Sleep(time.Minute * 5)
		break
	}
	// go Exit(cancele)
	// <-ctxend.Done()
	cancelu()
	cancels()
}
func Exit(cancel context.CancelFunc) {
	bufio.NewReader(os.Stdin).ReadBytes('q')
	cancel()
}

func Update(ctx context.Context, poolInterval int, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(poolInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Update()
		case <-ctx.Done():
			{
				//why I can't see this line in console?
				fmt.Println("ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
func Send(ctx context.Context, sendInterval int, url string, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(sendInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Send(url)
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
