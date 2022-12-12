package main

import (
	"bufio"
	"context"
	"fmt"
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
	fmt.Printf("%v", config)
	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper)
	ctxupdate, cancelu := context.WithCancel(context.Background())
	ctxsend, cancels := context.WithCancel(context.Background())
	//	ctxend, cancele := context.WithCancel(context.Background())
	go Update(ctxupdate, config.PollInterval, m)
	go Send(ctxsend, config.ReportInterval, m)

	defer cancelu()
	defer cancels()
	for {
		time.Sleep(time.Minute * 10)
		break
	}
	// go Exit(cancele)
	// <-ctxend.Done()

}
func Exit(cancel context.CancelFunc) {
	bufio.NewReader(os.Stdin).ReadBytes('q')
	cancel()
}

func Update(ctx context.Context, poolInterval int, service *memstatservice.SimpleMemStatService) {
	if poolInterval <= 0 {
		poolInterval = PollInterval
	}
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
func Send(ctx context.Context, sendInterval int, service *memstatservice.SimpleMemStatService) {
	if sendInterval <= 0 {
		sendInterval = ReportInterval
	}
	host := os.Getenv("ADDRESS")
	if host == "" {
		host = ServerURL
	}
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
