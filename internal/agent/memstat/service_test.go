package memstat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/stretchr/testify/assert"
)

const (
	fakeurl = "fake"
)

var (
	lock sync.Mutex
)

type simpleMockHTTPHelper struct {
	paths map[string]int
}

func (h *simpleMockHTTPHelper) Post(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error {
	m := models.Metrics{}
	json.Unmarshal(data.Bytes(), &m)
	formatedURL := ""
	if m.MType == gauge {
		formatedURL = fmt.Sprintf("%s/%s/", url, m.MType)
	} else {
		formatedURL = fmt.Sprintf("%s/%s/%s/%d", url, m.MType, m.ID, *m.Delta)
	}
	lock.Lock()
	defer lock.Unlock()
	for k := range h.paths {
		if strings.Contains(formatedURL, k) {
			h.paths[k]++
		}
	}
	return nil
}
func TestSimpleMemStatService_Update(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name     string
		values   runtime.MemStats
		want     runtime.MemStats
		wantPool int64
	}{
		{
			name:     "check send",
			values:   runtime.MemStats{},
			want:     runtime.MemStats{Alloc: 1234, Frees: 123},
			wantPool: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewService("")
			res.Update()
			assert.NotEqual(t, float64(res.gaugeDictionary["Alloc"]), 0, "alloc property not valid")
			assert.Equal(t, tt.wantPool, res.countDictionary[pollCount], "uint property  not valid")
		})
	}
}

func TestSimpleMemStatService_Send(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name      string
		values    map[string]float64
		sendCount int
		hhelper   simpleMockHTTPHelper
	}{
		{
			name:   "send some gauge",
			values: map[string]float64{"Alloc": 1234, "Frees": 123},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/":                      0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 2,
		},
		{
			name:   "send without gauge",
			values: map[string]float64{},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/":                      0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewService("")
			res.send = tt.hhelper.Post
			res.gaugeDictionary = tt.values
			res.Send(fakeurl)
			for k, v := range tt.hhelper.paths {
				if strings.Contains(k, gauge) {
					assert.Equal(t, tt.sendCount, v, fmt.Sprintf("%s should be called %d", k, tt.sendCount))
				} else {
					assert.Equal(t, 0, v, k+" should not be called")
				}
			}
		})
	}
}