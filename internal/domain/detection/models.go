// Package detection defines inefficiency detection domain models.
package detection

import "time"

// InefficiencyType categorizes detected inefficiencies.
type InefficiencyType string

const (
	TypeUnderutilizedGPU     InefficiencyType = "underutilized_gpu"
	TypeOversizedAllocation  InefficiencyType = "oversized_allocation"
	TypeLowMemoryUsage       InefficiencyType = "low_memory_usage"
	TypeCPUStarvation        InefficiencyType = "cpu_starvation"
	TypeStalledTraining      InefficiencyType = "stalled_training"
	TypeRepeatedFailures     InefficiencyType = "repeated_failures"
	TypeHighRetry            InefficiencyType = "high_retry_pattern"
	TypeAbnormalSpike        InefficiencyType = "abnormal_resource_spike"
	TypeIdleGPU              InefficiencyType = "idle_gpu"
	TypeDataLoaderBottleneck InefficiencyType = "data_loader_bottleneck"
	TypeXIDHardwareError     InefficiencyType = "xid_hardware_error"
	TypeCUDAMemoryFragmentation InefficiencyType = "cuda_memory_fragmentation"
	TypeNCCLStraggler        InefficiencyType = "nccl_straggler"
	TypeKVCacheSaturation    InefficiencyType = "kv_cache_saturation"
)

// Severity represents the severity of a detected inefficiency.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Inefficiency represents a detected inefficiency.
type Inefficiency struct {
	ID              string            `json:"id"`
	Type            InefficiencyType  `json:"type"`
	Severity        Severity          `json:"severity"`
	ClusterID       string            `json:"cluster_id"`
	NodeID          string            `json:"node_id"`
	GPUID           string            `json:"gpu_id,omitempty"`
	JobID           string            `json:"job_id,omitempty"`
	Description     string            `json:"description"`
	AffectedMetrics []string          `json:"affected_metrics"`
	CurrentValue    float64           `json:"current_value"`
	ThresholdValue  float64           `json:"threshold_value"`
	DetectedAt      time.Time         `json:"detected_at"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty"`
	IsActive        bool              `json:"is_active"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// DetectionRule defines a rule for detecting inefficiencies.
type DetectionRule struct {
	Name        string           `json:"name"`
	Type        InefficiencyType `json:"type"`
	Description string           `json:"description"`
	Enabled     bool             `json:"enabled"`
	Threshold   float64          `json:"threshold"`
	Duration    time.Duration    `json:"duration"` // How long condition must persist
	Severity    Severity         `json:"severity"`
	Cooldown    time.Duration    `json:"cooldown"` // Minimum time between alerts
}
