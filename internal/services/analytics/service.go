// Package analytics provides cluster-level analytics aggregation.
//
// This service periodically computes aggregated statistics like:
// - Cluster efficiency scores
// - Cost waste estimates
// - Utilization trends
// - Top-N underutilized GPUs
// - Job success/failure ratios
package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// ClusterStats holds computed analytics for a cluster.
type ClusterStats struct {
	ClusterID           string    `json:"cluster_id"`
	AvgGPUUtilization   float64   `json:"avg_gpu_utilization"`
	P50GPUUtilization   float64   `json:"p50_gpu_utilization"`
	P95GPUUtilization   float64   `json:"p95_gpu_utilization"`
	AvgMemoryUtilization float64  `json:"avg_memory_utilization"`
	IdleGPUCount        int       `json:"idle_gpu_count"`
	TotalGPUHours       float64   `json:"total_gpu_hours"`
	WastedGPUHours      float64   `json:"wasted_gpu_hours"`
	EstimatedWasteUSD   float64   `json:"estimated_waste_usd"`
	EfficiencyScore     float64   `json:"efficiency_score"` // 0-100
	ComputedAt          time.Time `json:"computed_at"`
}

// Service computes cluster analytics.
type Service struct {
	metricsRepo ports.MetricsRepository
	cache       ports.CacheStore
	log         *logger.Logger
}

// NewService creates a new analytics service.
func NewService(metricsRepo ports.MetricsRepository, cache ports.CacheStore, log *logger.Logger) *Service {
	return &Service{
		metricsRepo: metricsRepo,
		cache:       cache,
		log:         log.WithComponent("analytics"),
	}
}

// ComputeClusterStats calculates efficiency statistics for a cluster.
func (s *Service) ComputeClusterStats(ctx context.Context, clusterID string, window time.Duration) (*ClusterStats, error) {
	now := time.Now()
	from := now.Add(-window)

	gpuMetrics, err := s.metricsRepo.QueryGPUMetrics(ctx, ports.MetricsQuery{
		ClusterID: clusterID,
		From:      from,
		To:        now,
		Limit:     100000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	if len(gpuMetrics) == 0 {
		return &ClusterStats{ClusterID: clusterID, ComputedAt: now}, nil
	}

	// Compute aggregations
	var totalUtil, totalMem float64
	idleGPUs := make(map[string]bool)
	for _, m := range gpuMetrics {
		totalUtil += m.Utilization
		totalMem += m.MemoryUtilization
		if m.IsIdle(5.0) {
			idleGPUs[m.GPUID] = true
		}
	}

	count := float64(len(gpuMetrics))
	avgUtil := totalUtil / count
	avgMem := totalMem / count

	// Efficiency score: 100 = perfect, 0 = all idle
	efficiencyScore := avgUtil
	estimatedWaste := (100 - avgUtil) / 100 * count * 2.50 / 3600 // rough hourly estimate

	stats := &ClusterStats{
		ClusterID:            clusterID,
		AvgGPUUtilization:    avgUtil,
		AvgMemoryUtilization: avgMem,
		IdleGPUCount:         len(idleGPUs),
		EfficiencyScore:      efficiencyScore,
		EstimatedWasteUSD:    estimatedWaste,
		ComputedAt:           now,
	}

	// Cache the result
	s.cache.Set(ctx, fmt.Sprintf("cluster_stats:%s", clusterID), stats, 5*time.Minute)

	return stats, nil
}
