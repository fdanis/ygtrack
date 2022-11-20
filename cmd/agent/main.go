package main

import (
	"github.com/fdanis/ygtrack/cmd/agent/businesslayer/memstatservice"
	// "github.com/fdanis/ygtrack/internal/helpers/fakehttphelper"
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
	//memstatservice.HTTPHelper = fakehttphelper.Helper{}
	memStatS := memstatservice.NewMemStatService(gaugeList[:], "http://localhost:8080/update")
	memStatS.Run(2, 10)
}
