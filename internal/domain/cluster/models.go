// Package cluster defines GPU cluster domain models.
package cluster

import "time"

// Cluster represents a GPU compute cluster.
type Cluster struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"` // "gcp", "aws", "azure", "on-prem"
	Region    string            `json:"region"`
	Nodes     []Node            `json:"nodes"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	IsActive  bool              `json:"is_active"`
}

// Node represents a compute node in a cluster.
type Node struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"cluster_id"`
	MachineType string            `json:"machine_type"` // e.g., "a2-highgpu-8g"
	GPUType     string            `json:"gpu_type"`     // e.g., "NVIDIA A100-SXM4-80GB"
	GPUCount    int               `json:"gpu_count"`
	CPUCores    int               `json:"cpu_cores"`
	MemoryGB    float64           `json:"memory_gb"`
	Status      NodeStatus        `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// NodeStatus represents the operational status of a node.
type NodeStatus string

const (
	NodeStatusReady       NodeStatus = "ready"
	NodeStatusNotReady    NodeStatus = "not_ready"
	NodeStatusDraining    NodeStatus = "draining"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// CostConfig holds per-GPU cost information for a cluster.
type CostConfig struct {
	ClusterID      string  `json:"cluster_id"`
	GPUCostPerHour float64 `json:"gpu_cost_per_hour_usd"`
	CPUCostPerHour float64 `json:"cpu_cost_per_hour_usd"`
	Currency       string  `json:"currency"` // "USD"
}
