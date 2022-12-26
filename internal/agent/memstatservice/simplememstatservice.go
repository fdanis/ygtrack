package memstatservice

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/fdanis/ygtrack/internal/server/models"
)

const (
	gauge       = "gauge"
	counter     = "counter"
	pollCount   = "PollCount"
	randomCount = "RandomValue"
)

type SimpleMemStatService struct {
	gaugeDictionary map[string]float64
	pollCount       int64
	randomCount     int64
	send            func(url string, contentType string, data *bytes.Buffer) error
	lock            sync.RWMutex
}

func NewSimpleMemStatService() *SimpleMemStatService {
	m := new(SimpleMemStatService)
	m.send = post
	m.gaugeDictionary = map[string]float64{}
	return m
}
func (m *SimpleMemStatService) Update() {
	fmt.Printf("update metrics %s \n", time.Now().Format("15:04:05"))
	m.lock.Lock()
	defer m.lock.Unlock()
	memstat := runtime.MemStats{}
	runtime.ReadMemStats(&memstat)
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
	maxint := int64(^uint(0) >> 1)
	randomValue, err := rand.Int(rand.Reader, big.NewInt(maxint))
	if err != nil {
		panic(err)
	}
	m.randomCount = randomValue.Int64()
}

func (m *SimpleMemStatService) Send(url string) {
	fmt.Printf("send metrics %s \n", time.Now().Format("15:04:05"))
	copymap := make(map[string]float64, len(m.gaugeDictionary))
	m.lock.RLock()
	for key, val := range m.gaugeDictionary {
		copymap[key] = val
	}
	var poolCountValue = m.pollCount
	var randomCountValue = float64(m.randomCount)
	m.lock.RUnlock()

	for k, v := range copymap {
		//use go
		m.httpSendStat(&models.Metrics{ID: k, MType: gauge, Value: &v}, url)
	}
	//use go
	m.httpSendStat(&models.Metrics{ID: pollCount, MType: counter, Delta: &poolCountValue}, url)
	m.httpSendStat(&models.Metrics{ID: randomCount, MType: gauge, Value: &randomCountValue}, url)
}

func (m *SimpleMemStatService) httpSendStat(data *models.Metrics, url string) {
	//url := fmt.Sprintf("%s/%s/%s/%s", host, t, name, val)
	d, err := json.Marshal(data)

	if err != nil {
		log.Printf("could marshal %v", err)
	}
	err = m.send(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		log.Printf("could not send metric %v %v", data, err)
	}
}

func post(url string, contentType string, data *bytes.Buffer) error {
	res, err := http.Post(url, contentType, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got wrong http status (%d)", res.StatusCode)
	}
	return nil
}
