package main

import (
	"runtime"
	"time"

	"github.com/fdanis/ygtrack/cmd/agent/businesslayer/memstatservice"
	"github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
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
	hhelper := fakehttphelper.Helper{}
	memStatS := memstatservice.NewMemStatService(gaugeList[:], hhelper, runtime.ReadMemStats)
	Run(memStatS, 2, 10)
}

func Run(m *memstatservice.MemStatService[runtime.MemStats], pollInterval int, reportInterval int) {
	now := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		dur := time.Until(now)
		if int(dur.Seconds())%pollInterval == 0 {
			m.Update()
		}
		if int(dur.Seconds())%reportInterval == 0 {
			m.Send("http://localhost:8080/update")
		}
	}
}
