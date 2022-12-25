package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fdanis/ygtrack/internal/agent"
	"github.com/fdanis/ygtrack/internal/agent/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

func main() {
	config := agent.Conf{}
	agent.ReadFlags(&config)
	flag.Parse()
	agent.ReadEnv(&config)	
	m := memstatservice.NewSimpleMemStatService(httphelper.Post)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go Update(ctx, config.PollInterval, m)
	go Send(ctx, config.ReportInterval, config.Address, m)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig

	fmt.Println("exit")
}

func Update(ctx context.Context, poolInterval time.Duration, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(poolInterval)
	for {
		select {
		case <-t.C:
			service.Update()
		case <-ctx.Done():
			{
				fmt.Println("update ticker stoped")
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
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
