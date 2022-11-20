package memstatservice

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/helpers/httphelper"
)

var HTTPHelper helpers.HttpHelper = httphelper.Helper{}

type MemStatService struct {
	curent                runtime.MemStats
	metrics               []string
	reflectValue          map[string]reflect.Value
	url                   string
	pollCount             uint64
	randomCount           uint64
	r                     *rand.Rand
	secondsForUpdateTimer int
	secondsForSendTimer   int
}

func NewMemStatService(gaugelist []string, url string) *MemStatService {
	m := new(MemStatService)
	m.metrics = gaugelist
	m.url = url
	source := rand.NewSource(time.Now().UnixNano())
	m.r = rand.New(source)
	runtime.ReadMemStats(&m.curent)
	if m.reflectValue == nil {
		m.initReflection()
	}
	return m
}

func (m *MemStatService) Run(secondsForUpdateTimer int, secondsForSendTimer int) {
	now := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		dur := time.Until(now)
		if int(dur.Seconds())%secondsForUpdateTimer == 0 {
			m.Update()
		}
		if int(dur.Seconds())%secondsForSendTimer == 0 {
			m.Send()
		}
	}
}

func (m *MemStatService) Update() {
	fmt.Printf("update metrics %s \n", time.Now().Format("15:04:05"))
	runtime.ReadMemStats(&m.curent)
	m.pollCount++
	m.randomCount = m.r.Uint64()
}

func (m *MemStatService) Send() {
	fmt.Printf("send metrics %s \n", time.Now().Format("15:04:05"))
	for _, val := range m.metrics {
		go sendGaugeReflect(val, m.reflectValue[val], m.url)
	}
	go sendCounter("PollCount", m.pollCount, m.url)
	go sendGaugeNumber("RandomCount", m.randomCount, m.url)
}

func sendGaugeNumber(name string, val uint64, host string) {
	httpSendStat(name, "gauge", fmt.Sprintf("%d", val), host)
}

func sendGaugeReflect(name string, val reflect.Value, host string) {
	switch val.Type().Kind() {
	case reflect.Uint64:
		httpSendStat(name, "gauge", fmt.Sprintf("%d", val.Uint()), host)
	case reflect.Float64:
		httpSendStat(name, "gauge", fmt.Sprintf("%.f2", val.Float()), host)
	case reflect.Uint32:
		httpSendStat(name, "gauge", fmt.Sprintf("%d", val.Uint()), host)
	default:
		httpSendStat(name, "gauge", "-1", host)
	}
}

func sendCounter(name string, val uint64, host string) {
	httpSendStat(name, "counter", fmt.Sprintf("%d", val), host)
}

func httpSendStat(name string, t string, val string, host string) {
	url := fmt.Sprintf("%s/%s/%s/%s", host, t, name, val)
	err := HTTPHelper.Post(url)
	if err != nil {
		log.Printf("could not send %s metric with value %s %v", name, val, err)
	}
}

func (m *MemStatService) initReflection() {
	m.reflectValue = make(map[string]reflect.Value)
	r := reflect.ValueOf(m.curent)
	for _, val := range m.metrics {
		m.reflectValue[val] = reflect.Indirect(r).FieldByName(val)
	}
}
