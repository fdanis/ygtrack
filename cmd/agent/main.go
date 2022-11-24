package main

import (
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

var (
	gaugeList = [...]string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
)

func main() {
	hhelper := httphelper.Helper{}
	//hhelper := fakehttphelper.Helper{}
	m := memstatservice.NewMemStatService(gaugeList[:], hhelper, runtime.ReadMemStats)

	now := time.Now()
	t := time.NewTicker(1 * time.Second)
	for {
		<-t.C
		dur := time.Until(now)
		if int(dur.Seconds())%pollInterval == 0 {
			go m.Update()
		}
		if int(dur.Seconds())%reportInterval == 0 {
			go m.Send(serverURL)
		}
	}
}
