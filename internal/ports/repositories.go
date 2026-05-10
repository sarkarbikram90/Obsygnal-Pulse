// Package ports defines the interface boundaries (hexagonal architecture).
// These are the contracts that domain logic depends on.
// Adapters implement these interfaces for specific infrastructure.
package ports

import (
	"context"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/metrics"
)

// MetricsRepository defines the contract for metric data persistence.
type MetricsRepository interface {
	// InsertGPUMetrics batch-inserts GPU metrics.
	InsertGPUMetrics(ctx context.Context, metrics []metrics.GPUMetric) error
	// InsertNodeMetrics batch-inserts node metrics.
	InsertNodeMetrics(ctx context.Context, metrics []metrics.NodeMetric) error
	// QueryGPUMetrics retrieves GPU metrics within a time range.
	QueryGPUMetrics(ctx context.Context, query MetricsQuery) ([]metrics.GPUMetric, error)
	// QueryNodeMetrics retrieves node metrics within a time range.
	QueryNodeMetrics(ctx context.Context, query MetricsQuery) ([]metrics.NodeMetric, error)
	// GetGPUUtilizationAvg returns average GPU utilization over a time range.
	GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error)
	// GetClusterOverview returns aggregated cluster statistics.
	GetClusterOverview(ctx context.Context, clusterID string) (*ClusterOverview, error)
}

// MetricsQuery represents a flexible query for metrics.
type MetricsQuery struct {
	ClusterID string
	NodeID    string
	GPUID     string
	JobID     string
	From      time.Time
	To        time.Time
	Limit     int
	Offset    int
	OrderBy   string
	OrderDir  string // "asc" or "desc"
}

// ClusterOverview contains aggregated cluster statistics.
type ClusterOverview struct {
	ClusterID         string    `json:"cluster_id"`
	TotalNodes        int       `json:"total_nodes"`
	HealthyNodes      int       `json:"healthy_nodes"`
	TotalGPUs         int       `json:"total_gpus"`
	ActiveGPUs        int       `json:"active_gpus"`
	IdleGPUs          int       `json:"idle_gpus"`
	AvgUtilization    float64   `json:"avg_utilization"`
	AvgMemoryUsage    float64   `json:"avg_memory_usage"`
	ActiveJobs        int       `json:"active_jobs"`
	FailedJobs24h     int       `json:"failed_jobs_24h"`
	InefficiencyScore float64   `json:"inefficiency_score"`
	EstimatedWaste    float64   `json:"estimated_waste_usd"`
	Timestamp         time.Time `json:"timestamp"`
}

// MetricsCollector defines the contract for collecting metrics from external sources.
type MetricsCollector interface {
	// Collect retrieves current metrics from the source.
	Collect(ctx context.Context) (*metrics.MetricBatch, error)
	// Name returns the collector name for logging.
	Name() string
}

// MetricsPublisher defines the contract for publishing metrics to a queue.
type MetricsPublisher interface {
	// Publish sends a metric batch to the processing queue.
	Publish(ctx context.Context, batch *metrics.MetricBatch) error
	// Close cleanly shuts down the publisher.
	Close() error
}

// MetricsSubscriber defines the contract for consuming metrics from a queue.
type MetricsSubscriber interface {
	// Subscribe starts consuming metric batches, calling handler for each.
	Subscribe(ctx context.Context, handler func(ctx context.Context, batch *metrics.MetricBatch) error) error
	// Close cleanly shuts down the subscriber.
	Close() error
}

// CacheStore defines the contract for caching.
type CacheStore interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// Notifier defines the contract for sending notifications.
type Notifier interface {
	// Send dispatches a notification.
	Send(ctx context.Context, notification Notification) error
	// Type returns the notifier type (e.g., "slack", "webhook").
	Type() string
}

// Notification represents a message to send via a notification channel.
type Notification struct {
	Title    string            `json:"title"`
	Message  string            `json:"message"`
	Severity string            `json:"severity"` // "info", "warning", "critical"
	Fields   map[string]string `json:"fields,omitempty"`
	URL      string            `json:"url,omitempty"`
}
