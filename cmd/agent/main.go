package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/helpers"
	pb "github.com/fdanis/ygtrack/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

//go:generate go run ../generator/genvar.go string
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	config := agent.Conf{}
	f := agent.ReadFlags(&config)
	flag.Parse()
	err := agent.ReadEnv(&config)
	if err != nil {
		log.Fatal(err)
	}
	config.LoadFromConfigFile(*f)

	m := memstat.NewMetricService(config.Key)
	s := memstat.NewSenderMetric()

	if config.CryptoKey != "" {
		data, err := os.ReadFile(config.CryptoKey)
		if err != nil {
			panic("config file does not exists")
		}
		block, _ := pem.Decode([]byte(data))
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			panic(err)
		}

		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			log.Fatalf("got unexpected key type: %T", rsaKey)
		}
		s.PublicKey = rsaKey
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	printInfoVar()
	go Update(ctx, config.PollInterval, m)
	go UpdateGopsUtil(ctx, config.PollInterval, m)

	if !config.UseGrpc {
		go Send(ctx, config, m, s)
	} else {
		conn, err := grpc.Dial(config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		c := pb.NewMetricServiceClient(conn)
		go SendGrpc(ctx, config, m, s, c)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
	t := time.NewTicker(conf.ReportInterval)
	for {
		select {
		case <-t.C:
			{
				metrics := m.GetMetrics()
				if conf.Key != "" {
					err := helpers.SetHash(conf.Key, metrics)
					if err != nil {
						// don't send if error exists
						break
					}
				}
				log.Println("sending")
				s.Send("http://"+strings.TrimRight(conf.Address, "/")+"/update", metrics)
			}
		case <-ctx.Done():
			{
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}

func SendGrpc(ctx context.Context, conf agent.Conf, m *memstat.MetricService, s *memstat.SenderMetric, client pb.MetricServiceClient) {
	t := time.NewTicker(conf.ReportInterval)
	for {
		select {
		case <-t.C:
			{
				metrics := m.GetMetrics()
				if conf.Key != "" {
					err := helpers.SetHash(conf.Key, metrics)
					if err != nil {
						// don't send if error exists
						break
					}
				}
				log.Println("sending")
				stream, err := client.SendList(ctx)
				if err != nil {
					// don't send if error exists
					log.Println("can not get streaam")
				}

				for _, item := range metrics {
					p := &pb.Metrics{
						Id:    item.ID,
						MType: pb.Metrics_MetricsType(pb.Metrics_MetricsType_value[strings.ToUpper(item.MType)]),
						Hash:  item.Hash,
					}
					if item.MType == constants.MetricsTypeCounter {
						p.Delta = *item.Delta
					} else {
						p.Value = *item.Value
					}
					log.Printf("%v.Send(%v) = %v", stream, item, err)
					if err := stream.Send(p); err != nil {
						log.Printf("%v.Send(%v) = %v", stream, item, err)
					}

				}
				_, err = stream.CloseAndRecv()
				if err != nil {
					log.Printf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
				}
			}
		case <-ctx.Done():
			{
				fmt.Println("send ticker stoped")
				t.Stop()
				return
			}
		}
	}
}
