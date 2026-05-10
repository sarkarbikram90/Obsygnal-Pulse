// Package prometheus implements the MetricsCollector port using PromQL queries.
//
// This adapter queries Prometheus for DCGM Exporter metrics and converts them
// to our domain GPUMetric/NodeMetric models. It's the production replacement
// for the simulator.
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/metrics"
)

// Adapter collects GPU telemetry from Prometheus.
type Adapter struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewAdapter creates a new Prometheus metrics collector.
func NewAdapter(baseURL string, timeout time.Duration) *Adapter {
	return &Adapter{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
		timeout:    timeout,
	}
}

// Collect retrieves current GPU metrics from Prometheus via PromQL.
func (a *Adapter) Collect(ctx context.Context) (*metrics.MetricBatch, error) {
	// TODO: Implement PromQL queries for DCGM Exporter metrics:
	// - DCGM_FI_DEV_GPU_UTIL          → utilization
	// - DCGM_FI_DEV_MEM_COPY_UTIL     → memory_utilization
	// - DCGM_FI_DEV_GPU_TEMP          → temperature
	// - DCGM_FI_DEV_POWER_USAGE       → power_usage
	// - DCGM_FI_DEV_FB_USED           → memory_used
	// - DCGM_FI_DEV_FB_FREE           → memory_free
	// - DCGM_FI_DEV_PCIE_TX_THROUGHPUT → pcie_tx_bytes
	// - DCGM_FI_DEV_PCIE_RX_THROUGHPUT → pcie_rx_bytes
	return nil, fmt.Errorf("prometheus collector not yet implemented")
}

// Name returns the collector name.
func (a *Adapter) Name() string { return "prometheus" }

// QueryInstant executes an instant PromQL query.
func (a *Adapter) QueryInstant(ctx context.Context, query string) ([]PromResult, error) {
	// TODO: POST /api/v1/query with PromQL expression
	return nil, fmt.Errorf("not yet implemented")
}

// PromResult represents a single Prometheus query result.
type PromResult struct {
	Labels map[string]string
	Value  float64
	Time   time.Time
}
