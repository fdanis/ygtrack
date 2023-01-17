package memstatservice

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/server/models"
)

const (
	gauge       = "gauge"
	counter     = "counter"
	pollCount   = "PollCount"
	randomCount = "RandomValue"
)

type MemStatService struct {
	gaugeDictionary map[string]float64
	pollCount       int64
	randomCount     int64
	send            func(url string, header map[string]string, data *bytes.Buffer) error
	lock            sync.RWMutex
	hashkey         string
}

func NewMemStatService(hashkey string) *MemStatService {
	m := new(MemStatService)
	m.send = post
	m.gaugeDictionary = map[string]float64{}
	m.hashkey = hashkey
	return m
}
func (m *MemStatService) Update() {
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

func (m *MemStatService) Send(url string) {
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

func (m *MemStatService) SendBatch(url string) {
	fmt.Printf("send batch metrics %s \n", time.Now().Format("15:04:05"))
	m.lock.RLock()
	batch := make([]*models.Metrics, 0, len(m.gaugeDictionary)+2)
	for key, val := range m.gaugeDictionary {
		batch = append(batch, &models.Metrics{ID: key, MType: gauge, Value: &val})
	}
	var poolCountValue = m.pollCount
	var randomCountValue = float64(m.randomCount)
	batch = append(batch, &models.Metrics{ID: randomCount, MType: gauge, Value: &randomCountValue})
	batch = append(batch, &models.Metrics{ID: pollCount, MType: counter, Delta: &poolCountValue})
	m.lock.RUnlock()

	m.httpSendBatch(batch, url)
}

func (m *MemStatService) httpSendBatch(data []*models.Metrics, url string) {
	d, err := json.Marshal(data)
	if err != nil {
		log.Printf("could marshal %v", err)
		return
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)
	gz := helpers.GetPool().GetWriter(w)
	defer helpers.GetPool().PutWriter(gz)
	_, err = gz.Write(d)
	if err != nil {
		log.Println(err)
	}
	gz.Flush()

	err = m.send(url, map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"}, &buf)
	//	err = m.send(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		log.Printf("could not send metric %v %v to url %s", data, err, url)
	}
}

func (m *MemStatService) httpSendStat(data *models.Metrics, url string) {
	if m.hashkey != "" {
		if err := data.RefreshHash(m.hashkey); err != nil {
			log.Printf("could not refresh hash  %v", err)
		}
	}
	d, err := json.Marshal(data)
	if err != nil {
		log.Printf("could marshal %v", err)
	}
	err = m.send(url, map[string]string{"Content-Type": "application/json"}, bytes.NewBuffer(d))
	if err != nil {
		log.Printf("could not send metric %v %v", data, err)
	}
}

func post(url string, header map[string]string, data *bytes.Buffer) error {
	r, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}
	for k, v := range header {
		r.Header.Add(k, v)
	}

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got wrong http status (%d)", res.StatusCode)
	}
	return nil
}
