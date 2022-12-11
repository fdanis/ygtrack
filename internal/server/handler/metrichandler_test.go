package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/fdanis/ygtrack/internal/server/render"
	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository/metricrepository"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestMetricHandler_GetValue(t *testing.T) {
	type fields struct {
		counterStorage *map[string]dataclass.Metric[int64]
		gaugeStorage   *map[string]dataclass.Metric[float64]
	}

	type args struct {
		typeName   string
		metricName string
	}

	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Count"},
			want: want{
				code:        200,
				response:    "5",
				contentType: "",
			},
		},
		{
			name: "positive test minus counter",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: -2345}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Count"},
			want: want{
				code:        200,
				response:    "-2345",
				contentType: "",
			},
		},
		{
			name: "fake_type_counter should be 501 #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "fake_counter", metricName: "Count"},
			want: want{
				code:        501,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "incorect counter name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Fake_Count"},
			want: want{
				code:        404,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "incorect gouge name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "Fake_Count"},
			want: want{
				code:        404,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "get gouge by name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "test1"},
			want: want{
				code:        200,
				response:    "1.000",
				contentType: "",
			},
		},
		{
			name: "get gouge by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "Test1"},
			want: want{
				code:        200,
				response:    "1.000",
				contentType: "",
			},
		},
		{
			name: "get counter by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "counter", metricName: "COUNT"},
			want: want{
				code:        200,
				response:    "5",
				contentType: "",
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", tt.args.typeName, tt.args.metricName), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.args.typeName)
			rctx.URLParams.Add("name", tt.args.metricName)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			cr := metricrepository.NewMetricRepository[int64]()
			cr.Datastorage = tt.fields.counterStorage
			gr := metricrepository.NewMetricRepository[float64]()
			gr.Datastorage = tt.fields.gaugeStorage
			metricHandler := MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}

			h := http.HandlerFunc(metricHandler.GetValue)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestMetricHandler_GetAll(t *testing.T) {
	type fields struct {
		counterStorage *map[string]dataclass.Metric[int64]
		gaugeStorage   *map[string]dataclass.Metric[float64]
	}

	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			want: want{
				code:        200,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			var app config.AppConfig
			cachedTemplate, err := render.CreateTemplateCache()
			if err != nil {
				log.Fatal(err)
			}
			app.TemplateCache = cachedTemplate

			render.NewTemplates(&app)

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			w := httptest.NewRecorder()
			cr := metricrepository.NewMetricRepository[int64]()
			cr.Datastorage = tt.fields.counterStorage
			gr := metricrepository.NewMetricRepository[float64]()
			gr.Datastorage = tt.fields.gaugeStorage
			metricHandler := MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}

			h := http.HandlerFunc(metricHandler.Get)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotEmpty(t, resBody)

		})
	}
}

func TestMetricHandler_Update(t *testing.T) {
	type fields struct {
		counterStorage *map[string]dataclass.Metric[int64]
		gaugeStorage   *map[string]dataclass.Metric[float64]
	}

	type args struct {
		typeName   string
		metricName string
		value      string
	}

	// определяем структуру теста
	type want struct {
		code           int
		response       string
		contentType    string
		counterStorage []dataclass.Metric[int64]
		gaugeStorage   []dataclass.Metric[float64]
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		fields fields
		args   []args
		want   want
	}{
		// определяем все тесты
		{
			name: "add new metrics",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{},
			},
			args: []args{
				{typeName: "counter", metricName: "Count", value: "10"},
				{typeName: "counter", metricName: "Count", value: "20"},
				{typeName: "counter", metricName: "Count", value: "30"},
				{typeName: "gauge", metricName: "Test", value: "10.11"},
				{typeName: "gauge", metricName: "Test", value: "20.33"},
			},
			want: want{
				code:           200,
				response:       "",
				contentType:    "",
				counterStorage: []dataclass.Metric[int64]{{Name: "Count", Value: 60}},
				gaugeStorage:   []dataclass.Metric[float64]{{Name: "Test", Value: 20.33}},
			},
		},
		{
			name: "add incorect type metric",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{},
			},
			args: []args{
				{typeName: "Gcounter", metricName: "Count", value: "10"},
				{typeName: "Ggauge", metricName: "Test", value: "20.33"},
			},
			want: want{
				code:           501,
				response:       "",
				contentType:    "",
				counterStorage: []dataclass.Metric[int64]{},
				gaugeStorage:   []dataclass.Metric[float64]{},
			},
		},
		{
			name: "add incorect value metric",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{},
			},
			args: []args{
				{typeName: "counter", metricName: "Count", value: "45jhkj"},
				{typeName: "counter", metricName: "Count", value: "asdf20"},
				{typeName: "counter", metricName: "Count", value: "asdf"},
				{typeName: "counter", metricName: "Count", value: "34.34"},
				{typeName: "gauge", metricName: "Test", value: "erte"},
				{typeName: "gauge", metricName: "Test", value: "20fsd"},
				{typeName: "gauge", metricName: "Test", value: "asdf20fsd"},
			},
			want: want{
				code:           400,
				response:       "",
				contentType:    "",
				counterStorage: []dataclass.Metric[int64]{},
				gaugeStorage:   []dataclass.Metric[float64]{},
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			cr := metricrepository.NewMetricRepository[int64]()
			cr.Datastorage = tt.fields.counterStorage
			gr := metricrepository.NewMetricRepository[float64]()
			gr.Datastorage = tt.fields.gaugeStorage
			metricHandler := MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}

			for _, arg := range tt.args {

				request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/update/%s/%s/%s", arg.typeName, arg.metricName, arg.value), nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("type", arg.typeName)
				rctx.URLParams.Add("name", arg.metricName)
				rctx.URLParams.Add("value", arg.value)
				request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				h := http.HandlerFunc(metricHandler.Update)
				h.ServeHTTP(w, request)
				res := w.Result()

				if res.StatusCode != tt.want.code {
					t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
				}

				defer res.Body.Close()
				_, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
			}
			counterList := []dataclass.Metric[int64]{}
			for _, val := range *tt.fields.counterStorage {
				counterList = append(counterList, val)
			}
			gaugeList := []dataclass.Metric[float64]{}
			for _, val := range *tt.fields.gaugeStorage {
				gaugeList = append(gaugeList, val)
			}
			assert.ElementsMatch(t, gaugeList, tt.want.gaugeStorage)
			assert.ElementsMatch(t, counterList, tt.want.counterStorage)

		})
	}
}

func TestMetricHandler_GetValueJSON(t *testing.T) {
	type fields struct {
		counterStorage *map[string]dataclass.Metric[int64]
		gaugeStorage   *map[string]dataclass.Metric[float64]
	}

	type args struct {
		typeName   string
		metricName string
	}

	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Count"},
			want: want{
				code:        200,
				response:    "5",
				contentType: "application/json",
			},
		},
		{
			name: "positive test minus counter",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: -2345}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Count"},
			want: want{
				code:        200,
				response:    "-2345",
				contentType: "application/json",
			},
		},
		{
			name: "fake_type_counter should be 501 #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "fake_counter", metricName: "Count"},
			want: want{
				code:        501,
				response:    "",
				contentType: "application/json",
			},
		},
		{
			name: "incorect counter name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}},
			},
			args: args{typeName: "counter", metricName: "Fake_Count"},
			want: want{
				code:        200,
				response:    "",
				contentType: "application/json",
			},
		},
		{
			name: "incorect gouge name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "Fake_Count"},
			want: want{
				code:        200,
				response:    "",
				contentType: "application/json",
			},
		},
		{
			name: "get gouge by name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "test1"},
			want: want{
				code:        200,
				response:    "1.000",
				contentType: "application/json",
			},
		},
		{
			name: "get gouge by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "Test1"},
			want: want{
				code:        200,
				response:    "1.000",
				contentType: "application/json",
			},
		},
		{
			name: "get counter by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[int64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "counter", metricName: "COUNT"},
			want: want{
				code:        200,
				response:    "5",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			model := models.Metrics{ID: tt.args.metricName, MType: tt.args.typeName}
			data, err := json.Marshal(model)
			if err != nil {
				log.Fatal()
			}
			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(data))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			cr := metricrepository.NewMetricRepository[int64]()
			cr.Datastorage = tt.fields.counterStorage
			gr := metricrepository.NewMetricRepository[float64]()
			gr.Datastorage = tt.fields.gaugeStorage
			metricHandler := MetricHandler{CounterRepo: &cr, GaugeRepo: &gr}

			h := http.HandlerFunc(metricHandler.GetJSONValue)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()

			dec := json.NewDecoder(res.Body)
			dec.Decode(&model)

			if err != nil {
				t.Fatal(err)
			}

			switch tt.args.typeName {
			case "gauge":

				if s, err := strconv.ParseFloat(tt.want.response, 64); err == nil {
					assert.Equal(t, s, *model.Value)
				} else {
					assert.Nil(t, model.Value, "value should be empty")
				}
			case "counter":
				if s, err := strconv.ParseInt(tt.want.response, 10, 64); err == nil {
					assert.Equal(t, s, *model.Delta)
				} else {
					assert.Nil(t, model.Delta, "value should be empty")
				}

			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
