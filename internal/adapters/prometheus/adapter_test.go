package prometheus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrometheusAdapter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PromAPIResponse{
			Status: "success",
		}
		resp.Data.ResultType = "vector"
		resp.Data.Result = []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		}{
			{
				Metric: map[string]string{
					"gpu_id":     "gpu-01",
					"cluster_id": "c1",
					"instance":   "n1",
				},
				Value: []interface{}{12345678, "75.5"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, time.Second)
	assert.Equal(t, "prometheus", adapter.Name())

	batch, err := adapter.Collect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, batch)
	assert.NotEmpty(t, batch.GPUMetrics)
}

func TestPrometheusAdapter_Defaults(t *testing.T) {
	adapter := NewAdapter("", 0)
	assert.Equal(t, "http://localhost:9090", adapter.baseURL)
	assert.Equal(t, 10*time.Second, adapter.timeout)
}
