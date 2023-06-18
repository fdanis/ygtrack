package models

import (
	"testing"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SetHash(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name string
		m    *Metrics
		args args
	}{
		{
			name: "testHashFunction",
			m:    &Metrics{},
			args: args{hash: "hash"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.SetHash(tt.args.hash)
			assert.Equal(t, tt.args.hash, tt.m.Hash)
		})
	}
}

var (
	delta int64   = 10
	value float64 = 5
)

func TestMetrics_String(t *testing.T) {
	tests := []struct {
		name string
		m    Metrics
		want string
	}{
		{
			name: "testString",
			m: Metrics{MType: constants.MetricsTypeCounter,
				Delta: &delta,
				Value: &value,
				ID:    "1"},
			want: "1:" + constants.MetricsTypeCounter + ":10",
		},
		{
			name: "testString",
			m: Metrics{MType: constants.MetricsTypeGauge,
				Delta: &delta,
				Value: &value,
				ID:    "1"},
			want: "1:" + constants.MetricsTypeGauge + ":5.000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("Metrics.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		m       *Metrics
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name: "testUnmarshalGauge",
			m:    &Metrics{},
			args: args{[]byte(`{ "id":"1", "type":"gauge", "value":5}`)},
			want: &Metrics{ID: "1"},
		},
		{
			name: "testUnmarshalCount",
			m:    &Metrics{},
			args: args{[]byte(`{ "id":"1", "type":"count", "delta":5}`)},
			want: &Metrics{ID: "1"},
		},
		{
			name:    "testUnmarshalerror",
			m:       &Metrics{},
			args:    args{[]byte(`{ sdf"id":"1", "type":"count", "delta":5}`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want.ID, tt.m.ID)
			}
		})
	}
}
