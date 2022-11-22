package memstatservice

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/fdanis/ygtrack/internal/helpers"
)

type MemStatService[T any] struct {
	curent       T
	metrics      []string
	reflectValue map[string]reflect.Value
	pollCount    uint64
	randomCount  uint64
	r            *rand.Rand
	httpHelper   helpers.HTTPHelper
	updatefunc   func(obj *T)
}

const (
	gauge       = "gauge"
	counter     = "counter"
	pollCount   = "PollCount"
	randomCount = "RandomCount"
)

func NewMemStatService[T any](gaugelist []string, hhelp helpers.HTTPHelper, statupdate func(obj *T)) *MemStatService[T] {
	m := new(MemStatService[T])
	m.metrics = gaugelist
	source := rand.NewSource(time.Now().UnixNano())
	m.r = rand.New(source)
	m.updatefunc = statupdate
	m.httpHelper = hhelp
	m.updatefunc(&m.curent)
	m.initReflection()
	return m
}
func (m *MemStatService[T]) initReflection() {
	m.reflectValue = make(map[string]reflect.Value)
	r := reflect.ValueOf(&m.curent)

	for _, val := range m.metrics {
		field := reflect.Indirect(r).FieldByName(val)
		if field.IsValid() {
			m.reflectValue[val] = field
		}
	}
}

func (m *MemStatService[T]) Update() {
	fmt.Printf("update metrics %s \n", time.Now().Format("15:04:05"))
	m.updatefunc(&m.curent)
	m.pollCount++
	m.randomCount = m.r.Uint64()
}

func (m *MemStatService[T]) Send(url string) {
	fmt.Printf("send metrics %s \n", time.Now().Format("15:04:05"))
	for _, val := range m.metrics {
		if r, ok := m.reflectValue[val]; ok {
			//use go
			m.httpSendStat(val, gauge, getReflectValue(r), url)
		}
	}
	//use go
	m.httpSendStat(pollCount, counter, fmt.Sprintf("%d", m.pollCount), url)
	m.httpSendStat(randomCount, gauge, fmt.Sprintf("%d", m.randomCount), url)
}

func getReflectValue(val reflect.Value) string {
	var v string
	switch val.Type().Kind() {
	case reflect.Uint64:
		v = fmt.Sprintf("%d", val.Uint())
	case reflect.Float64:
		v = fmt.Sprintf("%.2f", val.Float())
	case reflect.Uint32:
		v = fmt.Sprintf("%d", val.Uint())
	default:
		v = "-1"
	}
	return v
}

func (m *MemStatService[T]) httpSendStat(name string, t string, val string, host string) {
	url := fmt.Sprintf("%s/%s/%s/%s", host, t, name, val)
	err := m.httpHelper.Post(url)
	if err != nil {
		log.Printf("could not send %s metric with value %s %v", name, val, err)
	}
}
