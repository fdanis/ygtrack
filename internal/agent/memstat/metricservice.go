package memstat

import (
	"crypto/rand"
	"log"
	"math/big"
	"runtime"
	"sync/atomic"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	pollCount   = "PollCount"
	randomCount = "RandomValue"
)

type MetricService struct {
	gaugeDictionary map[string]float64
	countValue      int64
	allmetrics      []*models.Metrics
	hashkey         string
	w               chan int
	w1              chan int
	w2              chan int
	r               chan int
}

func NewMetricService(hashkey string) *MetricService {
	m := &MetricService{
		gaugeDictionary: make(map[string]float64, 31),
		hashkey:         hashkey,
		r:               make(chan int),
		w:               make(chan int),
		w1:              make(chan int),
		w2:              make(chan int),
	}
	resetChan(m.r)
	resetChan(m.w)
	resetChan(m.w1)
	resetChan(m.w2)
	return m
}

func resetChan(ch chan int) {
	go func(ch chan int) { ch <- 1 }(ch)
}
func (m *MetricService) Update() {
	atomic.AddInt64(&m.countValue, 1)
	select {
	case <-m.w1:
		<-m.w
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
		m.setAllMetrics()
		resetChan(m.w)
		resetChan(m.w1)
	default:
		return
	}
}

func (m *MetricService) UpdateGopsUtil() {
	select {
	case <-m.w2:
		<-m.w
		virtual, err := mem.VirtualMemory()
		if err != nil {
			log.Println(err)
		}
		m.gaugeDictionary["FreeMemory"] = float64(virtual.Free)
		m.gaugeDictionary["TotalMemory"] = float64(virtual.Total)
		m.gaugeDictionary["CPUutilization1"] = float64(virtual.UsedPercent)
		m.setAllMetrics()
		resetChan(m.w)
		resetChan(m.w2)
	default:
		return
	}
}

func (m *MetricService) setAllMetrics() {
	allmetrics := make([]*models.Metrics, 0, len(m.gaugeDictionary)+1)
	for key, val := range m.gaugeDictionary {
		v := val
		allmetrics = append(allmetrics, &models.Metrics{ID: key, MType: constants.MetricsTypeGauge, Value: &v})
	}
	m.allmetrics = allmetrics
}

func (m *MetricService) GetMetrics() []*models.Metrics {
	if m.allmetrics == nil {
		return nil
	}
	return append(m.allmetrics, &models.Metrics{ID: pollCount, MType: constants.MetricsTypeCounter, Delta: &m.countValue})
}
