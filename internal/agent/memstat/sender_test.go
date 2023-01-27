package memstat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/fdanis/ygtrack/internal/constants"
	"github.com/fdanis/ygtrack/internal/server/models"
	"github.com/stretchr/testify/assert"
)

const (
	fakeurl = "fake"
)

var (
	lock sync.Mutex
)

type simpleMockHTTPHelper struct {
	paths map[string]int
}

func (h *simpleMockHTTPHelper) Post(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error {
	m := models.Metrics{}
	json.Unmarshal(data.Bytes(), &m)
	formatedURL := ""
	if m.MType == constants.MetricsTypeGauge {
		formatedURL = fmt.Sprintf("%s/%s/", url, m.MType)
	} else {
		formatedURL = fmt.Sprintf("%s/%s/%s/%d", url, m.MType, m.ID, *m.Delta)
	}
	lock.Lock()
	defer lock.Unlock()
	for k := range h.paths {
		if strings.Contains(formatedURL, k) {
			h.paths[k]++
		}
	}
	return nil
}

func TestSimpleMemStatService_Send(t *testing.T) {
	tests := []struct { // добавился слайс тестов
		name      string
		values    []*models.Metrics
		sendCount int
		hhelper   simpleMockHTTPHelper
	}{
		{
			name: "send some gauge",
			values: []*models.Metrics{
				{ID: "Alloc", MType: constants.MetricsTypeGauge, Value: func(i float64) *float64 { return &i }(1234)},
				{ID: "Frees", MType: constants.MetricsTypeGauge, Value: func(i float64) *float64 { return &i }(123)},
			},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + constants.MetricsTypeGauge + "/":                      0,
				fakeurl + "/" + constants.MetricsTypeCounter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 2,
		},
		{
			name:   "send without gauge",
			values: []*models.Metrics{},
			hhelper: simpleMockHTTPHelper{paths: map[string]int{
				fakeurl + "/" + constants.MetricsTypeGauge + "/":                      0,
				fakeurl + "/" + constants.MetricsTypeCounter + "/" + pollCount + "/1": 0,
			}},
			sendCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewSenderMetric()
			res.send = tt.hhelper.Post

			res.Send(fakeurl, tt.values)
			for k, v := range tt.hhelper.paths {
				if strings.Contains(k, constants.MetricsTypeGauge) {
					assert.Equal(t, tt.sendCount, v, fmt.Sprintf("%s should be called %d", k, tt.sendCount))
				} else {
					assert.Equal(t, 0, v, k+" should not be called")
				}
			}
		})
	}
}
