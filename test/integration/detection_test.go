package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	detectionsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/detection"
)

// mockMetricsRepo mocks ClickHouse query results for testing detection rules
type mockMetricsRepo struct {
	gpuMetrics  []metrics.GPUMetric
	nodeMetrics []metrics.NodeMetric
}

func (m *mockMetricsRepo) QueryGPUMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.GPUMetric, error) {
	return m.gpuMetrics, nil
}

func (m *mockMetricsRepo) QueryNodeMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.NodeMetric, error) {
	return m.nodeMetrics, nil
}
func (m *mockMetricsRepo) GetClusterOverview(ctx context.Context, clusterID string) (*ports.ClusterOverview, error) {
	return &ports.ClusterOverview{}, nil
}
func (m *mockMetricsRepo) GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error) {
	return 0.0, nil
}
func (m *mockMetricsRepo) InsertBatch(ctx context.Context, batch []metrics.GPUMetric) error {
	return nil
}
func (m *mockMetricsRepo) InsertGPUMetrics(ctx context.Context, metrics []metrics.GPUMetric) error {
	return nil
}
func (m *mockMetricsRepo) InsertNodeMetrics(ctx context.Context, metrics []metrics.NodeMetric) error {
	return nil
}

type mockIneffRepo struct {
	inserted []detection.Inefficiency
}

func (m *mockIneffRepo) InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error {
	m.inserted = append(m.inserted, ineffs...)
	return nil
}
func (m *mockIneffRepo) Insert(ctx context.Context, ineff *detection.Inefficiency) error { return nil }
func (m *mockIneffRepo) GetByID(ctx context.Context, id string) (*detection.Inefficiency, error) {
	return nil, nil
}
func (m *mockIneffRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error) {
	return nil, nil
}
func (m *mockIneffRepo) ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error) {
	return nil, nil
}
func (m *mockIneffRepo) Resolve(ctx context.Context, id string, resolvedAt time.Time) error {
	return nil
}
func (m *mockIneffRepo) CountActive(ctx context.Context, clusterID string) (int, error) {
	return 0, nil
}

type mockCache struct{}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *mockCache) Delete(ctx context.Context, key string) error                { return nil }
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error)        { return false, nil }
func (m *mockCache) Incr(ctx context.Context, key string) (int64, error)         { return 0, nil }
func (m *mockCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return true, nil
}

func TestDetectionRules(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping detection integration test. Set INTEGRATION_TEST=1 to run.")
	}

	log := logger.New("development")

	now := time.Now()

	// Create synthetic telemetry data representing an Idle GPU
	gpuMetrics := []metrics.GPUMetric{}
	for i := 0; i < 15; i++ {
		gpuMetrics = append(gpuMetrics, metrics.GPUMetric{
			ID:          "metric-1",
			Timestamp:   now.Add(time.Duration(-i) * time.Minute),
			ClusterID:   "cluster-test",
			NodeID:      "node-1",
			GPUID:       "gpu-1",
			Utilization: 1.0, // under 2.0 threshold for idle
			MemoryUsed:  1000000,
			MemoryTotal: 80000000000,
		})
	}

	metricsRepo := &mockMetricsRepo{gpuMetrics: gpuMetrics}
	ineffRepo := &mockIneffRepo{}
	cache := &mockCache{}

	// Load default config
	fromConfig := config.DetectionConfig{
		EvaluationInterval:         60 * time.Second,
		GPUUtilizationThreshold:    20.0,
		MemoryUtilizationThreshold: 15.0,
	}

	svc := detectionsvc.NewService(metricsRepo, ineffRepo, cache, log, fromConfig)

	// Register a listener
	var detected []detection.Inefficiency
	svc.OnDetection(func(ctx context.Context, ineff detection.Inefficiency) {
		detected = append(detected, ineff)
	})

	// Run evaluation loop once manually
	_, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(detected) == 0 {
		t.Errorf("Expected idle GPU inefficiency to be detected")
	} else {
		if detected[0].Type != detection.TypeIdleGPU {
			t.Errorf("Expected TypeIdleGPU, got %s", detected[0].Type)
		}
	}
}
