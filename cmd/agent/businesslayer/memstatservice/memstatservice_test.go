package memstatservice

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMemStat struct {
	UInt64   uint64
	Float    float64
	UInt32   uint32
	WasCount uint
}

func mock_read_stat(stat *mockMemStat) {
	stat.WasCount++
	stat.UInt64 = uint64(stat.WasCount) * 1
	stat.Float = float64(stat.WasCount) * 1.1
	stat.UInt32 = uint32(stat.WasCount) * 3
}

const (
	fakeurl = "fake"
)

type mockHttpHelper struct {
	paths map[string]int
}

func (h *mockHttpHelper) Get(url string) error {
	return nil
}

func (h *mockHttpHelper) Post(url string) error {
	fmt.Println(url)
	for k, _ := range h.paths {
		if strings.Contains(url, k) {
			h.paths[k]++
		}
	}
	return nil
}

func TestMemStatService_New(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name     string
		values   mockMemStat
		want     mockMemStat
		wantPool uint64
	}{
		{
			name:     "check initialization",
			values:   mockMemStat{},
			want:     mockMemStat{UInt64: 1, Float: 1.1, UInt32: 3, WasCount: 1},
			wantPool: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewMemStatService[mockMemStat]([]string{"Float", "IncorectParam"}, nil, mock_read_stat)
			assert.Equal(t, res.curent.UInt64, tt.want.UInt64, "uint property was not sent")
			assert.Equal(t, res.curent.Float, tt.want.Float, "uint property was not sent")
			assert.Equal(t, res.curent.UInt32, tt.want.UInt32, "uint property was not sent")
			assert.Equal(t, res.curent.WasCount, tt.want.WasCount, "uint property was not sent")
			assert.Equal(t, res.pollCount, tt.wantPool, "pool should not be set")
			assert.Equal(t, len(res.reflectValue), 1, "reflect values should be 1")
		})
	}
}

func TestMemStatService_Update(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name     string
		values   mockMemStat
		want     mockMemStat
		wantPool uint64
	}{
		{
			name:     "check send",
			values:   mockMemStat{},
			want:     mockMemStat{UInt64: 2, Float: 2.2, UInt32: 6, WasCount: 2},
			wantPool: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewMemStatService[mockMemStat]([]string{"First"}, nil, mock_read_stat)
			res.Update()
			assert.Equal(t, res.curent.UInt64, tt.want.UInt64, "uint property not valid")
			assert.Equal(t, res.curent.Float, tt.want.Float, "uint property  not valid")
			assert.Equal(t, res.curent.UInt32, tt.want.UInt32, "uint property not valid")
			assert.Equal(t, res.curent.WasCount, tt.want.WasCount, "uint property not valid")
			assert.Equal(t, res.pollCount, tt.wantPool, "pool count incorect")
		})
	}
}

func TestMemStatService_Send(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name     string
		values   mockMemStat
		want     mockMemStat
		wantPool uint64
		hhelper  mockHttpHelper
	}{
		{
			name:   "check update",
			values: mockMemStat{},
			want:   mockMemStat{UInt64: 1, Float: 2.2, UInt32: 3, WasCount: 1},
			hhelper: mockHttpHelper{paths: map[string]int{
				fakeurl + "/" + gauge + "/" + "Float/2.20":       0,
				fakeurl + "/" + counter + "/" + pollCount + "/1": 0,
				fakeurl + "/" + gauge + "/" + randomCount + "/":  0,
			}},
			wantPool: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewMemStatService[mockMemStat]([]string{"Float"}, &tt.hhelper, mock_read_stat)
			res.Update()
			res.Send(fakeurl)
			for k, v := range tt.hhelper.paths {
				assert.Equal(t, v, 1, k+" should be called once")
			}
		})
	}
}
