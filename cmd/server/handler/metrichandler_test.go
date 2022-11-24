package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fdanis/ygtrack/cmd/server/store/dataclass"
	"github.com/fdanis/ygtrack/cmd/server/store/repository/metricrepository"
	"github.com/go-chi/chi"
)

func TestMetricHandler_GetValue(t *testing.T) {
	type fields struct {
		counterStorage *map[string]dataclass.Metric[uint64]
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
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
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
			name: "fake_type_counter should be 501 #1",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
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
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
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
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
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
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "test1"},
			want: want{
				code:        200,
				response:    "1.00",
				contentType: "",
			},
		},
		{
			name: "get gouge by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
				gaugeStorage:   &map[string]dataclass.Metric[float64]{"test1": {Name: "TEst1", Value: 1}, "test2": {Name: "test2", Value: 2}}},
			args: args{typeName: "gauge", metricName: "Test1"},
			want: want{
				code:        200,
				response:    "1.00",
				contentType: "",
			},
		},
		{
			name: "get counter by upercase name",
			fields: fields{
				counterStorage: &map[string]dataclass.Metric[uint64]{"count": {Name: "Count", Value: 5}},
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
			cr := metricrepository.NewMetricRepository[uint64]()
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
