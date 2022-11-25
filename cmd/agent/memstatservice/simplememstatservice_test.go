package memstatservice

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

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

func (h *simpleMockHTTPHelper) Post(url string) error {
	fmt.Println(url)
	for k := range h.paths {
		if strings.Contains(url, k) {
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
			res := NewSimpleMemStatService(nil, mockRuntimeReadStat)
			res.Update()
			assert.Equal(t, float64(res.gaugeDictionary["Alloc"]), float64(tt.want.Alloc), "uint property not valid")
			assert.Equal(t, float64(res.gaugeDictionary["Frees"]), float64(tt.want.Frees), "uint property  not valid")
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
				fakeurl + "/" + gauge + "/" + "Alloc/1234":       0,
				fakeurl + "/" + gauge + "/" + "Frees/123":        0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
				fakeurl + "/" + gauge + "/" + randomCount + "/":  0,
			}},
			wantPool: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewSimpleMemStatService(&tt.hhelper, mockRuntimeReadStat)
			res.Update()
			res.Send(fakeurl)
			for k, v := range tt.hhelper.paths {
				assert.Equal(t, v, 1, k+" should be called once")
			}
		})
	}
}
