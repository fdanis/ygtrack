package memstat

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"runtime"
	"sync"
	"time"

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
	countDictionary map[string]int64
	lock            sync.RWMutex
	hashkey         string
}

func NewMetricService(hashkey string) *MetricService {
	m := new(MetricService)
	m.gaugeDictionary = map[string]float64{}
	m.countDictionary = map[string]int64{pollCount: 0}
	m.hashkey = hashkey
	return m
}
func (m *MetricService) Update() {
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

func (m *MetricService) UpdateGopsUtil() {
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

func (m *MetricService) GetMetrics() []*models.Metrics {
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
