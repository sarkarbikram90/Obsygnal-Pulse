// Package jobs defines training job domain models.
package jobs

import "time"

// Status represents the lifecycle state of a training job.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
	StatusOOM       Status = "oom" // Out of memory
)

// Job represents a training job running on the cluster.
type Job struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	ClusterID   string            `json:"cluster_id"`
	NodeID      string            `json:"node_id"`
	Status      Status            `json:"status"`
	GPUIDs      []string          `json:"gpu_ids"`
	GPUCount    int               `json:"gpu_count"`
	Framework   string            `json:"framework"` // pytorch, tensorflow, jax
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Runtime     time.Duration     `json:"runtime"`
	QueueWait   time.Duration     `json:"queue_wait"`
	RetryCount  int               `json:"retry_count"`
	ExitCode    int               `json:"exit_code"`
	ErrorMsg    string            `json:"error_msg,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`

	// Resource allocation
	RequestedGPUs   int     `json:"requested_gpus"`
	RequestedMemory uint64  `json:"requested_memory"` // bytes
	RequestedCPUs   float64 `json:"requested_cpus"`

	// Utilization during job
	AvgGPUUtilization float64 `json:"avg_gpu_utilization"`
	MaxGPUUtilization float64 `json:"max_gpu_utilization"`
	AvgMemoryUsage    float64 `json:"avg_memory_usage"`
	MaxMemoryUsage    float64 `json:"max_memory_usage"`
}

// IsActive returns true if the job is currently running.
func (j *Job) IsActive() bool {
	return j.Status == StatusRunning || j.Status == StatusPending
}

// IsFailed returns true if the job ended in failure.
func (j *Job) IsFailed() bool {
	return j.Status == StatusFailed || j.Status == StatusOOM
}

// IsUnderutilized returns true if the job's average GPU utilization is below threshold.
func (j *Job) IsUnderutilized(threshold float64) bool {
	return j.AvgGPUUtilization < threshold && j.Status == StatusRunning
}

// HasExcessiveRetries returns true if retry count exceeds threshold.
func (j *Job) HasExcessiveRetries(threshold int) bool {
	return j.RetryCount > threshold
}
