package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/stretchr/testify/assert"
)

type mockMetricsRepo struct {
	metrics []metrics.GPUMetric
}

func (m *mockMetricsRepo) QueryGPUMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.GPUMetric, error) {
	return m.metrics, nil
}
func (m *mockMetricsRepo) QueryNodeMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.NodeMetric, error) {
	return nil, nil
}
func (m *mockMetricsRepo) GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error) {
	return 0, nil
}
func (m *mockMetricsRepo) GetClusterOverview(ctx context.Context, clusterID string) (*ports.ClusterOverview, error) {
	return nil, nil
}
func (m *mockMetricsRepo) InsertGPUMetrics(ctx context.Context, metrics []metrics.GPUMetric) error {
	return nil
}
func (m *mockMetricsRepo) InsertNodeMetrics(ctx context.Context, metrics []metrics.NodeMetric) error {
	return nil
}

type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockCache) Delete(ctx context.Context, key string) error         { return nil }
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }

func TestAnalyticsService(t *testing.T) {
	now := time.Now()
	repo := &mockMetricsRepo{
		metrics: []metrics.GPUMetric{
			{
				GPUID:       "g1",
				Utilization: 80.0,
				MemoryUsed:  40,
				MemoryTotal: 80,
				Timestamp:   now,
			},
			{
				GPUID:            "g2",
				Utilization:      0.0,
				RunningProcesses: 0,
				Timestamp:        now,
			},
		},
	}

	log := logger.New("test")
	svc := NewService(repo, &mockCache{}, log)

	stats, err := svc.ComputeClusterStats(context.Background(), "c1", time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, "c1", stats.ClusterID)
	assert.Equal(t, 40.0, stats.AvgGPUUtilization)
	assert.Equal(t, 1, stats.IdleGPUCount)
}
