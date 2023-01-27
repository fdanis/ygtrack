package memstat

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			res := NewMetricService("")
			res.Update()
			assert.NotEqual(t, float64(res.gaugeDictionary["Alloc"]), 0, "alloc property not valid")
			assert.Equal(t, tt.wantPool, res.countDictionary[pollCount], "uint property  not valid")
		})
	}
}
