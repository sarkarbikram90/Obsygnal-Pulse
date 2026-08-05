package memory

import (
	"context"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_MetricsRepository(t *testing.T) {
	log := logger.New("test")
	store := NewRepositoryStore(log)
	ctx := context.Background()

	now := time.Now()
	gpuMetric := metrics.GPUMetric{
		ID:          "gpu-test-1",
		Timestamp:   now,
		ClusterID:   "c1",
		NodeID:      "n1",
		GPUID:       "g1",
		Utilization: 75.0,
	}
	nodeMetric := metrics.NodeMetric{
		Timestamp: now,
		ClusterID: "c1",
		NodeID:    "n1",
		CPUUsage:  45.0,
	}

	assert.NoError(t, store.InsertGPUMetrics(ctx, []metrics.GPUMetric{gpuMetric}))
	assert.NoError(t, store.InsertNodeMetrics(ctx, []metrics.NodeMetric{nodeMetric}))

	gpuResults, err := store.QueryGPUMetrics(ctx, ports.MetricsQuery{ClusterID: "c1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, gpuResults)

	nodeResults, err := store.QueryNodeMetrics(ctx, ports.MetricsQuery{ClusterID: "c1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, nodeResults)

	avgUtil, err := store.GetGPUUtilizationAvg(ctx, "g1", now.Add(-time.Hour), now.Add(time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, 75.0, avgUtil)

	overview, err := store.GetClusterOverview(ctx, "c1")
	assert.NoError(t, err)
	assert.Equal(t, "c1", overview.ClusterID)
}

func TestMemoryStore_InefficiencyRepository(t *testing.T) {
	log := logger.New("test")
	store := NewRepositoryStore(log)
	repo := store.InefficiencyRepo()
	ctx := context.Background()

	ineff := &detection.Inefficiency{
		ID:        "ineff-test-01",
		Type:      detection.TypeIdleGPU,
		Severity:  detection.SeverityHigh,
		ClusterID: "c1",
		NodeID:    "n1",
		IsActive:  true,
	}

	assert.NoError(t, repo.Insert(ctx, ineff))

	fetched, err := repo.GetByID(ctx, "ineff-test-01")
	assert.NoError(t, err)
	assert.Equal(t, detection.TypeIdleGPU, fetched.Type)

	activeList, err := repo.ListActive(ctx, "c1", 10, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, activeList)

	count, err := repo.CountActive(ctx, "c1")
	assert.NoError(t, err)
	assert.True(t, count > 0)

	assert.NoError(t, repo.Resolve(ctx, "ineff-test-01", time.Now()))
}

func TestMemoryStore_RecommendationRepository(t *testing.T) {
	log := logger.New("test")
	store := NewRepositoryStore(log)
	repo := store.RecommendationRepo()
	ctx := context.Background()

	rec := &recommendation.Recommendation{
		ID:        "rec-test-01",
		Category:  recommendation.CategoryReduceIdleTime,
		Status:    recommendation.StatusActive,
		ClusterID: "c1",
		NodeID:    "n1",
	}

	assert.NoError(t, repo.Insert(ctx, rec))

	fetched, err := repo.GetByID(ctx, "rec-test-01")
	assert.NoError(t, err)
	assert.Equal(t, recommendation.CategoryReduceIdleTime, fetched.Category)

	activeList, err := repo.ListActive(ctx, "c1", 10, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, activeList)

	count, err := repo.CountActive(ctx, "c1")
	assert.NoError(t, err)
	assert.True(t, count > 0)

	assert.NoError(t, repo.Apply(ctx, "rec-test-01"))
	assert.NoError(t, repo.Dismiss(ctx, "rec-test-01", "user", "test dismissal"))
}

func TestMemoryStore_AlertRepository(t *testing.T) {
	log := logger.New("test")
	store := NewRepositoryStore(log)
	repo := store.AlertRepo()
	ctx := context.Background()

	alert := &ports.Alert{
		ID:        "alert-test-01",
		Type:      "idle_gpu",
		Severity:  "high",
		ClusterID: "c1",
	}

	assert.NoError(t, repo.Insert(ctx, alert))

	recent, err := repo.ListRecent(ctx, "c1", 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, recent)

	assert.NoError(t, repo.Acknowledge(ctx, "alert-test-01", "operator"))
}
