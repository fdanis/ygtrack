package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fdanis/ygtrack/cmd/agent/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
	//"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
)

const (
	pollInterval   int    = 2
	reportInterval int    = 10
	serverURL      string = "http://localhost:8080/update"
)

func main() {
	hhelper := httphelper.Helper{}
	m := memstatservice.NewSimpleMemStatService(hhelper, runtime.ReadMemStats)

	ctxupdate, cancelu := context.WithCancel(context.Background())
	ctxsend, cancels := context.WithCancel(context.Background())
	//	ctxend, cancele := context.WithCancel(context.Background())
	go Update(ctxupdate, pollInterval, m)
	go Send(ctxsend, reportInterval, m)

	for {

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
func Send(ctx context.Context, sendInterval int, service *memstatservice.SimpleMemStatService) {
	t := time.NewTicker(time.Duration(sendInterval) * time.Second)
	for {
		select {
		case <-t.C:
			service.Send(serverURL)
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
