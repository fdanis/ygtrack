package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fdanis/ygtrack/internal/agent"
	"github.com/fdanis/ygtrack/internal/agent/memstat"
	"github.com/fdanis/ygtrack/internal/helpers"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

func main() {
	config := agent.Conf{}
	agent.ReadFlags(&config)
	flag.Parse()
	err := agent.ReadEnv(&config)
	if err != nil {
		log.Fatal(err)
	}
	m := memstat.NewMetricService(config.Key)
	s := memstat.NewSenderMetric()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go Update(ctx, config.PollInterval, m)
	go UpdateGopsUtil(ctx, config.PollInterval, m)
	go Send(ctx, config, m, s)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig

	fmt.Println("exit")
}

func Update(ctx context.Context, poolInterval time.Duration, service *memstat.MetricService) {
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
func UpdateGopsUtil(ctx context.Context, poolInterval time.Duration, service *memstat.MetricService) {
	t := time.NewTicker(poolInterval)
	for {
		select {
		case <-t.C:
			service.UpdateGopsUtil()
		case <-ctx.Done():
			{
				fmt.Println("update ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
func Send(ctx context.Context, conf agent.Conf, m *memstat.MetricService, s *memstat.SenderMetric) {
	t := time.NewTicker(time.Duration(conf.ReportInterval.Seconds()))
	for {
		select {
		case <-t.C:
			metrics := m.GetMetrics()
			if conf.Key != "" {
				err := helpers.SetHash(conf.Key, metrics)
				if err != nil {
					// don't send if error exists
					break
				}
			}
			s.Send("http://"+strings.TrimRight(conf.Address, "/")+"/update", metrics)
		case <-ctx.Done():
			{
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
