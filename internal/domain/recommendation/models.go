// Package recommendation defines optimization recommendation domain models.
package recommendation

import "time"

// Category categorizes the type of optimization recommendation.
type Category string

const (
	CategoryReduceAllocation   Category = "reduce_gpu_allocation"
	CategoryIncreaseBatchSize  Category = "increase_batch_size"
	CategoryIncreaseWorkers    Category = "increase_dataloader_workers"
	CategoryMoveToSmallerNode  Category = "move_to_smaller_node"
	CategoryInvestigateCPU     Category = "investigate_cpu_bottleneck"
	CategoryInvestigateMemory  Category = "investigate_memory_fragmentation"
	CategoryReduceIdleTime     Category = "reduce_idle_runtime"
	CategoryOptimizeCheckpoint Category = "optimize_checkpointing"
	CategoryRightSizeMemory    Category = "right_size_memory"
)

// Status tracks the lifecycle of a recommendation.
type Status string

const (
	StatusActive    Status = "active"
	StatusDismissed Status = "dismissed"
	StatusApplied   Status = "applied"
	StatusExpired   Status = "expired"
)

// Recommendation represents an actionable optimization suggestion.
type Recommendation struct {
	ID              string            `json:"id"`
	Category        Category          `json:"category"`
	Status          Status            `json:"status"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	Explanation     string            `json:"explanation"`
	ConfidenceScore float64           `json:"confidence_score"` // 0.0 - 1.0
	EstimatedSavings EstimatedSavings `json:"estimated_savings"`
	AffectedMetrics []string          `json:"affected_metrics"`
	ClusterID       string            `json:"cluster_id"`
	NodeID          string            `json:"node_id,omitempty"`
	GPUID           string            `json:"gpu_id,omitempty"`
	JobID           string            `json:"job_id,omitempty"`
	InefficiencyID  string            `json:"inefficiency_id"`
	CreatedAt       time.Time         `json:"created_at"`
	ExpiresAt       time.Time         `json:"expires_at"`
	DismissedAt     *time.Time        `json:"dismissed_at,omitempty"`
	DismissedBy     string            `json:"dismissed_by,omitempty"`
	DismissReason   string            `json:"dismiss_reason,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// EstimatedSavings represents the estimated impact of applying a recommendation.
type EstimatedSavings struct {
	CostPerHourUSD    float64 `json:"cost_per_hour_usd"`
	CostPerMonthUSD   float64 `json:"cost_per_month_usd"`
	GPUHoursPerMonth  float64 `json:"gpu_hours_per_month"`
	MemoryGB          float64 `json:"memory_gb,omitempty"`
	PowerWatts        float64 `json:"power_watts,omitempty"`
}
