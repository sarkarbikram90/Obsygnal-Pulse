// Package metrics defines the core domain models representing GPU telemetry data.
// These models encapsulate the purest business rules of the platform (e.g., defining
// what an "idle" GPU means or parsing threshold validations) and are completely
// isolated from any external infrastructure, database ORMs, or HTTP frameworks.
package metrics

import (
	"fmt"
	"time"
)

// GPUMetric represents a single GPU telemetry data point.
type GPUMetric struct {
	// Identity
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ClusterID string    `json:"cluster_id"`
	NodeID    string    `json:"node_id"`
	GPUID     string    `json:"gpu_id"`
	GPUModel  string    `json:"gpu_model"`
	GPUIndex  int       `json:"gpu_index"`

	// Utilization metrics (0-100%)
	Utilization       float64 `json:"utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`

	// Memory metrics (bytes)
	MemoryUsed  uint64 `json:"memory_used"`
	MemoryTotal uint64 `json:"memory_total"`
	MemoryFree  uint64 `json:"memory_free"`

	// Thermal and power
	Temperature float64 `json:"temperature"` // Celsius
	PowerUsage  float64 `json:"power_usage"` // Watts
	PowerLimit  float64 `json:"power_limit"` // Watts
	FanSpeed    float64 `json:"fan_speed"`   // Percentage

	// PCIe metrics
	PCIeTxBytes uint64 `json:"pcie_tx_bytes"` // bytes/sec
	PCIeRxBytes uint64 `json:"pcie_rx_bytes"` // bytes/sec

	// NVIDIA Hardware Primitives & Interconnects
	MIGInstanceID     string  `json:"mig_instance_id,omitempty"`     // MIG Slice ID (e.g. 1g.10gb)
	NVLinkBandwidthGB float64 `json:"nvlink_bandwidth_gbps"`          // Inter-GPU P2P throughput
	XIDErrorCode      int     `json:"xid_error_code,omitempty"`       // Critical Hardware Fault Code (e.g., 31, 43, 62, 79)
	TensorCoreUtil    float64 `json:"tensor_core_utilization"`        // Matrix multiplier activity

	// Deep AI Training & LLM Inference Profiler
	CUDAMemReserved   uint64  `json:"cuda_memory_reserved_bytes"`   // PyTorch memory_reserved()
	KVCacheUtilization float64 `json:"kv_cache_utilization,omitempty"` // vLLM PagedAttention KV Cache %
	TokensPerSecond   float64 `json:"tokens_per_second,omitempty"`    // LLM inference throughput

	// Process info
	RunningProcesses int    `json:"running_processes"`
	JobID            string `json:"job_id,omitempty"`

	// Labels for flexible filtering
	Labels map[string]string `json:"labels,omitempty"`
}

// Validate checks that a GPUMetric has valid values.
func (m *GPUMetric) Validate() error {
	if m.ClusterID == "" {
		return fmt.Errorf("cluster_id is required")
	}
	if m.NodeID == "" {
		return fmt.Errorf("node_id is required")
	}
	if m.GPUID == "" {
		return fmt.Errorf("gpu_id is required")
	}
	if m.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	if m.Utilization < 0 || m.Utilization > 100 {
		return fmt.Errorf("utilization must be 0-100, got: %f", m.Utilization)
	}
	if m.MemoryUtilization < 0 || m.MemoryUtilization > 100 {
		return fmt.Errorf("memory_utilization must be 0-100, got: %f", m.MemoryUtilization)
	}
	if m.Temperature < -50 || m.Temperature > 150 {
		return fmt.Errorf("temperature out of range: %f", m.Temperature)
	}
	return nil
}

// IsIdle returns true if the GPU is considered idle.
func (m *GPUMetric) IsIdle(threshold float64) bool {
	return m.Utilization < threshold && m.RunningProcesses == 0
}

// IsOverheating returns true if GPU temperature exceeds threshold.
func (m *GPUMetric) IsOverheating(threshold float64) bool {
	return m.Temperature > threshold
}

// MemoryUsagePercent returns memory usage as a percentage.
func (m *GPUMetric) MemoryUsagePercent() float64 {
	if m.MemoryTotal == 0 {
		return 0
	}
	return float64(m.MemoryUsed) / float64(m.MemoryTotal) * 100
}

// NodeMetric represents node-level resource metrics.
type NodeMetric struct {
	Timestamp   time.Time         `json:"timestamp"`
	ClusterID   string            `json:"cluster_id"`
	NodeID      string            `json:"node_id"`
	CPUUsage    float64           `json:"cpu_usage"`    // 0-100%
	MemoryUsage float64           `json:"memory_usage"` // 0-100%
	DiskUsage   float64           `json:"disk_usage"`   // 0-100%
	NetworkIn   uint64            `json:"network_in"`   // bytes/sec
	NetworkOut  uint64            `json:"network_out"`  // bytes/sec
	GPUCount    int               `json:"gpu_count"`
	PodCount    int               `json:"pod_count"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// MetricBatch represents a batch of metrics for efficient processing.
type MetricBatch struct {
	GPUMetrics  []GPUMetric  `json:"gpu_metrics"`
	NodeMetrics []NodeMetric `json:"node_metrics"`
	ReceivedAt  time.Time    `json:"received_at"`
	Source      string       `json:"source"` // "prometheus", "simulator"
}

// Size returns the total number of metrics in the batch.
func (b *MetricBatch) Size() int {
	return len(b.GPUMetrics) + len(b.NodeMetrics)
}
