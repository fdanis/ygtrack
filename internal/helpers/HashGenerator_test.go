package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockHashedObject struct {
	hash string
}

func (m *MockHashedObject) SetHash(hash string) {
	m.hash = hash
}

func TestSetHash(t *testing.T) {
	type args struct {
		key      string
		datalist []*MockHashedObject
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "sethash",
			args: args{key: "secret", datalist: []*MockHashedObject{{}, {}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetHash(tt.args.key, tt.args.datalist); (err != nil) != tt.wantErr {
				t.Errorf("SetHash() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, v := range tt.args.datalist {
				assert.NotEqual(t, "", v.hash)
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	type args struct {
		data any
		key  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "getHash",
			args: args{data: "sometext", key: "secret"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetHash(tt.args.data, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotEqual(t, "", got)
		})
	}
}
