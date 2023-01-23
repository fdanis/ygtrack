package memstat

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

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
)

const (
	pollCount   = "PollCount"
	randomCount = "RandomValue"
)

type Service struct {
	gaugeDictionary map[string]float64
	countDictionary map[string]int64
	send            func(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error
	lock            sync.RWMutex
	hashkey         string
	httpclient      *http.Client
	workers         int
}

func NewService(hashkey string) *Service {
	m := new(Service)
	m.send = post
	m.httpclient = &http.Client{}
	m.gaugeDictionary = map[string]float64{}
	m.countDictionary = map[string]int64{pollCount: 0}
	m.hashkey = hashkey
	m.workers = 10
	return m
}
func (m *Service) Update() {
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

	maxint := int64(^uint(0) >> 1)
	randomValue, err := rand.Int(rand.Reader, big.NewInt(maxint))
	if err != nil {
		panic(err)
	}
	m.gaugeDictionary[randomCount] = float64(randomValue.Int64())
	m.countDictionary[pollCount]++
}

func (m *Service) UpdateGopsUtil() {
	fmt.Printf("update gops util metrics %s \n", time.Now().Format("15:04:05"))
	m.lock.Lock()
	defer m.lock.Unlock()
	virtual, err := mem.VirtualMemory()
	if err != nil {
		log.Println(err)
	}
	m.gaugeDictionary["FreeMemory"] = float64(virtual.Free)
	m.gaugeDictionary["TotalMemory"] = float64(virtual.Total)
	m.gaugeDictionary["CPUutilization1"] = float64(virtual.UsedPercent)
}

func (m *Service) Send(url string) {
	metrics := m.getMetrics()
	//generate hash
	if m.hashkey != "" {
		g := &errgroup.Group{}
		for _, v := range metrics {
			gen := helpers.HashGenerator{Metric: v, Key: m.hashkey}
			g.Go(gen.Do)
		}
		err := g.Wait()
		if err != nil {
			log.Printf("could not refresh hash  %v", err)
			//don't send any metrics if exists error
			return
		}
	}

	g := errgroup.Group{}
	recordCh := make(chan *bytes.Buffer)
	for i := 0; i < m.workers; i++ {
		w := &SendWorker{
			ch: recordCh,
			send: func(data *bytes.Buffer) error {
				return m.send(m.httpclient, url, map[string]string{"Content-Type": "application/json"}, data)
			},
		}
		g.Go(w.Do)
	}

	w := &MarshalWorker{ch: recordCh, list: metrics}
	err := w.Do()
	if err != nil {
		log.Println(err)
	}
	close(recordCh)

	err = g.Wait()
	if err != nil {
		log.Println(err)
	}
}

type MarshalWorker struct {
	list []*models.Metrics
	ch   chan *bytes.Buffer
}

func (w *MarshalWorker) Do() error {
	for _, item := range w.list {
		d, err := json.Marshal(item)
		if err != nil {
			log.Printf("could marshal %v", err)
			return err
		}
		w.ch <- bytes.NewBuffer(d)
	}
	return nil
}

type SendWorker struct {
	ch   chan *bytes.Buffer
	send func(data *bytes.Buffer) error
}

func (w *SendWorker) Do() error {
	for data := range w.ch {
		err := w.send(data)
		if err != nil {
			log.Print(err)
		}
	}
	return nil
}

func (m *Service) SendBatch(url string) {
	metrics := m.getMetrics()
	d, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("could not do json.marshal %v", err)
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

	err = post(m.httpclient, url, map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"}, &buf)
	if err != nil {
		log.Println("can not send batch")
	}
}

func (m *Service) getMetrics() []*models.Metrics {
	m.lock.RLock()
	allmetrics := make([]*models.Metrics, 0, len(m.gaugeDictionary)+2)
	for key, val := range m.gaugeDictionary {
		v := val
		allmetrics = append(allmetrics, &models.Metrics{ID: key, MType: constants.MetricsTypeGauge, Value: &v})
	}
	for key, val := range m.countDictionary {
		v := val
		allmetrics = append(allmetrics, &models.Metrics{ID: key, MType: constants.MetricsTypeCounter, Delta: &v})
	}
	m.lock.RUnlock()
	return allmetrics
}

func post(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error {
	r, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}
	for k, v := range header {
		r.Header.Add(k, v)
	}
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
