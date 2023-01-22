package memstatservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/stretchr/testify/assert"
)

const (
	fakeurl = "fake"
)

type simpleMockHTTPHelper struct {
	paths map[string]int
}

func (h *simpleMockHTTPHelper) Post(url string, header map[string]string, data *bytes.Buffer) error {
	m := models.Metrics{}
	json.Unmarshal(data.Bytes(), &m)
	formatedURL := ""
	if m.MType == gauge {
		formatedURL = fmt.Sprintf("%s/%s/", url, m.MType)
	} else {
		formatedURL = fmt.Sprintf("%s/%s/%s/%d", url, m.MType, m.ID, *m.Delta)
	}
	println(formatedURL)
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
			res := NewMemStatService("")
			res.Update()
			assert.NotEqual(t, float64(res.gaugeDictionary["Alloc"]), 0, "alloc property not valid")
			assert.Equal(t, res.pollCount, tt.wantPool, "uint property  not valid")
		})
	}
}

func TestSimpleMemStatService_Send(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name      string
		values    map[string]float64
		sendCount uint64
		hhelper   simpleMockHTTPHelper
	}{
		{
			name:   "send some gauge",
			values: map[string]float64{"Alloc": 1234, "Frees": 123},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/":                      0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 3,
		},
		{
			name:   "send without gauge",
			values: map[string]float64{},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/":                      0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewMemStatService("")
			res.send = tt.hhelper.Post
			res.gaugeDictionary = tt.values
			res.Send(fakeurl)
			for k, v := range tt.hhelper.paths {
				if strings.Contains(k, gauge) {
					assert.Equal(t, v, tt.sendCount, fmt.Sprintf("%s should be called %d", k, tt.sendCount))
				} else {
					assert.Equal(t, v, 1, k+" should be called once")
				}
			}
		})
	}
}
