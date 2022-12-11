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

func mockRuntimeReadStat(stat *runtime.MemStats) {
	//stat.WasCount++
	stat.Alloc = 1234
	stat.Frees = 123
}

type simpleMockHTTPHelper struct {
	paths map[string]int
}

func (h *simpleMockHTTPHelper) Get(url string) error {
	return nil
}

func (h *simpleMockHTTPHelper) Post(url string, contentType string, data *bytes.Buffer) error {
	m := models.Metrics{}
	json.Unmarshal(data.Bytes(), &m)
	formatedUrl := ""
	if m.MType == gauge {
		formatedUrl = fmt.Sprintf("%s/%s/", url, m.MType)
	} else {

		formatedUrl = fmt.Sprintf("%s/%s/%s/%d", url, m.MType, m.ID, *m.Delta)
	}
	println(formatedUrl)
	for k := range h.paths {
		if strings.Contains(formatedUrl, k) {
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
			res := NewSimpleMemStatService(nil)
			res.Update()
			assert.NotEqual(t, float64(res.gaugeDictionary["Alloc"]), 0, "alloc property not valid")
			assert.Equal(t, res.pollCount, tt.wantPool, "uint property  not valid")
		})
	}
}

func TestSimpleMemStatService_Send(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name     string
		values   runtime.MemStats
		want     runtime.MemStats
		wantPool uint64
		hhelper  simpleMockHTTPHelper
	}{
		{
			name:   "check update",
			values: runtime.MemStats{},
			want:   runtime.MemStats{Alloc: 1234, Frees: 123},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/":                      0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
			}},
			wantPool: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewSimpleMemStatService(&tt.hhelper)
			res.Update()
			res.Send(fakeurl)
			for k, v := range tt.hhelper.paths {
				if strings.Contains(k, gauge) {
					assert.Equal(t, v, 28, k+" should be called once")
				} else {
					assert.Equal(t, v, 1, k+" should be called once")
				}
			}
		})
	}
}
