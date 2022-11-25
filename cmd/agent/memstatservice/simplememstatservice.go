package memstatservice

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/fdanis/ygtrack/internal/helpers"
)

type SimpleMemStatService struct {
	gaugeDictionary map[string]float64
	pollCount       int64
	randomCount     int64
	r               *rand.Rand
	httpHelper      helpers.HTTPHelper
	updatefunc      func(obj *runtime.MemStats)
	lock            sync.RWMutex
}

func NewSimpleMemStatService(hhelp helpers.HTTPHelper, statupdate func(obj *runtime.MemStats)) *SimpleMemStatService {
	m := new(SimpleMemStatService)
	source := rand.NewSource(time.Now().UnixNano())
	m.r = rand.New(source)
	m.updatefunc = statupdate
	m.httpHelper = hhelp
	m.gaugeDictionary = map[string]float64{}
	return m
}
func (m *SimpleMemStatService) Update() {
	fmt.Printf("update metrics %s \n", time.Now().Format("15:04:05"))
	m.lock.Lock()
	defer m.lock.Unlock()
	memstat := runtime.MemStats{}
	m.updatefunc(&memstat)
	m.gaugeDictionary["Alloc"] = float64(memstat.Alloc)
	m.gaugeDictionary["BuckHashSys"] = float64(memstat.BuckHashSys)
	m.gaugeDictionary["Frees"] = float64(memstat.Frees)
	m.gaugeDictionary["GCCPUFraction"] = float64(memstat.GCCPUFraction)
	m.gaugeDictionary["GCSys"] = float64(memstat.GCSys)
	m.gaugeDictionary["HeapAlloc"] = float64(memstat.HeapAlloc)
	m.gaugeDictionary["HeapIdle"] = float64(memstat.HeapIdle)
	m.gaugeDictionary["HeapInuse"] = float64(memstat.HeapInuse)
	m.gaugeDictionary["HeapObjects"] = float64(memstat.HeapObjects)
	m.gaugeDictionary["HeapReleased"] = float64(memstat.HeapReleased)
	m.gaugeDictionary["HeapSys"] = float64(memstat.HeapSys)
	m.gaugeDictionary["LastGC"] = float64(memstat.LastGC)
	m.gaugeDictionary["Lookups"] = float64(memstat.Lookups)
	m.gaugeDictionary["MCacheInuse"] = float64(memstat.MCacheInuse)
	m.gaugeDictionary["MCacheSys"] = float64(memstat.MCacheSys)
	m.gaugeDictionary["MSpanInuse"] = float64(memstat.MSpanInuse)
	m.gaugeDictionary["MSpanSys"] = float64(memstat.MSpanSys)
	m.gaugeDictionary["Mallocs"] = float64(memstat.Mallocs)
	m.gaugeDictionary["NextGC"] = float64(memstat.NextGC)
	m.gaugeDictionary["NumForcedGC"] = float64(memstat.NumForcedGC)
	m.gaugeDictionary["NumGC"] = float64(memstat.NumGC)
	m.gaugeDictionary["OtherSys"] = float64(memstat.OtherSys)
	m.gaugeDictionary["PauseTotalNs"] = float64(memstat.PauseTotalNs)
	m.gaugeDictionary["StackInuse"] = float64(memstat.StackInuse)
	m.gaugeDictionary["StackSys"] = float64(memstat.StackSys)
	m.gaugeDictionary["Sys"] = float64(memstat.Sys)
	m.gaugeDictionary["TotalAlloc"] = float64(memstat.TotalAlloc)

	m.pollCount++
	m.randomCount = m.r.Int63()
}

func (m *SimpleMemStatService) Send(url string) {
	fmt.Printf("send metrics %s \n", time.Now().Format("15:04:05"))
	copymap := make(map[string]float64, len(m.gaugeDictionary))
	m.lock.RLock()
	for key, val := range m.gaugeDictionary {
		copymap[key] = val
	}
	var poolCountValue = m.pollCount
	var randomCountValue = m.randomCount
	m.lock.RUnlock()
	for k, v := range copymap {
		//use go
		m.httpSendStat(k, gauge, fmt.Sprintf("%.3f", v), url)
	}
	//use go
	m.httpSendStat(pollCount, counter, fmt.Sprintf("%d", poolCountValue), url)
	m.httpSendStat(randomCount, gauge, fmt.Sprintf("%d", randomCountValue), url)
}

func (m *SimpleMemStatService) httpSendStat(name string, t string, val string, host string) {
	url := fmt.Sprintf("%s/%s/%s/%s", host, t, name, val)
	err := m.httpHelper.Post(url)
	if err != nil {
		log.Printf("could not send %s metric with value %s %v", name, val, err)
	}
}
