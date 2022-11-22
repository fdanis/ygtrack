package counterrepository

import (
	"reflect"
	"testing"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
	"github.com/stretchr/testify/assert"
)

func TestCounterRepository_Add(t *testing.T) {
	type fields struct {
		Datastorage *map[string]uint64
	}
	type args struct {
		data dataclass.CounterMetric
	}
	tests := []struct {
		name    string
		fields  fields
		args    []args
		wantErr bool
	}{
		{
			name:    "test add without datastorage",
			fields:  fields{},
			args:    []args{{data: dataclass.CounterMetric{Name: "TestName", Value: 2}}},
			wantErr: true,
		},
		{
			name:    "should be added one element",
			fields:  fields{Datastorage: &map[string]uint64{}},
			args:    []args{{data: dataclass.CounterMetric{Name: "TestName", Value: 1}}},
			wantErr: false,
		},
		{
			name:   "should be added 2 elements",
			fields: fields{Datastorage: &map[string]uint64{}},
			args: []args{
				{data: dataclass.CounterMetric{Name: "TestName", Value: 1}},
				{data: dataclass.CounterMetric{Name: "TestName", Value: 2}},
			},
			wantErr: false,
		},
		{
			name:   "should be added 2 elements in 2 types",
			fields: fields{Datastorage: &map[string]uint64{}},
			args: []args{
				{data: dataclass.CounterMetric{Name: "TestName", Value: 1}},
				{data: dataclass.CounterMetric{Name: "TestName", Value: 2}},
				{data: dataclass.CounterMetric{Name: "Test3Name", Value: 1999}},
				{data: dataclass.CounterMetric{Name: "Test3Name", Value: 23333}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CounterRepository{Datastorage: tt.fields.Datastorage}
			typ := map[string]uint64{}

			for _, v := range tt.args {
				if err := r.Add(v.data); (err != nil) != tt.wantErr {
					t.Errorf("CounterRepository.Add() error = %v, wantErr %v", err, tt.wantErr)
				}
				typ[v.data.Name] += v.data.Value
			}
			if !tt.wantErr {
				assert.Equal(t, len(*r.Datastorage), len(typ))
				for k, v := range typ {
					assert.Equal(t, (*r.Datastorage)[k], v)
				}
			}
		})
	}
}

func TestCounterRepository_GetByName(t *testing.T) {
	type fields struct {
		Datastorage *map[string]uint64
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dataclass.CounterMetric
		wantErr bool
	}{
		{
			name:    "Get by name",
			fields:  fields{Datastorage: &map[string]uint64{"Test1": 1, "Test2": 2}},
			args:    args{name: "Test1"},
			want:    &dataclass.CounterMetric{Name: "Test1", Value: 1},
			wantErr: false,
		},
		{
			name:    "Get without datastorage",
			fields:  fields{Datastorage: nil},
			args:    args{name: "Test1"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Get by not existing name",
			fields:  fields{Datastorage: &map[string]uint64{"Test": 1, "Test2": 2}},
			args:    args{name: "Test3"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CounterRepository{
				Datastorage: tt.fields.Datastorage,
			}
			got, err := r.GetByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CounterRepository.GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil && tt.want != nil {
				t.Errorf("CounterRepository.GetByName() = %v, want %v", got, *tt.want)
				return
			}
			if got == nil && tt.want == nil {
				return
			}
			if !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("CounterRepository.GetByName() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestCounterRepository_GetAll(t *testing.T) {
	type fields struct {
		Datastorage *map[string]uint64
	}
	tests := []struct {
		name    string
		fields  fields
		want    []dataclass.CounterMetric
		wantErr bool
	}{
		{
			name:   "get all",
			fields: fields{Datastorage: &map[string]uint64{"Test1": 1, "Test2": 2, "Test3": 3}},
			want: []dataclass.CounterMetric{
				dataclass.CounterMetric{Name: "Test1", Value: 1},
				dataclass.CounterMetric{Name: "Test2", Value: 2},
				dataclass.CounterMetric{Name: "Test3", Value: 3},
			},
			wantErr: false,
		},
		{
			name:    "get all",
			fields:  fields{Datastorage: nil},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CounterRepository{
				Datastorage: tt.fields.Datastorage,
			}
			got, err := r.GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("CounterRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.ElementsMatch(t, got, tt.want)
			}
		})
	}
}
