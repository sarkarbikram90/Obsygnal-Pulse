package detection

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// ---- Mock Implementations ----

type mockMetricsRepo struct {
	gpuMetrics  []metrics.GPUMetric
	nodeMetrics []metrics.NodeMetric
}

func (m *mockMetricsRepo) InsertGPUMetrics(ctx context.Context, gm []metrics.GPUMetric) error {
	return nil
}
func (m *mockMetricsRepo) InsertNodeMetrics(ctx context.Context, nm []metrics.NodeMetric) error {
	return nil
}
func (m *mockMetricsRepo) QueryGPUMetrics(ctx context.Context, q ports.MetricsQuery) ([]metrics.GPUMetric, error) {
	return m.gpuMetrics, nil
}
func (m *mockMetricsRepo) QueryNodeMetrics(ctx context.Context, q ports.MetricsQuery) ([]metrics.NodeMetric, error) {
	return m.nodeMetrics, nil
}
func (m *mockMetricsRepo) GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error) {
	return 50.0, nil
}
func (m *mockMetricsRepo) GetClusterOverview(ctx context.Context, clusterID string) (*ports.ClusterOverview, error) {
	return &ports.ClusterOverview{}, nil
}

type mockIneffRepo struct {
	mu     sync.Mutex
	ineffs []detection.Inefficiency
}

func (m *mockIneffRepo) Insert(ctx context.Context, ineff *detection.Inefficiency) error { return nil }
func (m *mockIneffRepo) InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ineffs = append(m.ineffs, ineffs...)
	return nil
}
func (m *mockIneffRepo) GetByID(ctx context.Context, id string) (*detection.Inefficiency, error) {
	return nil, nil
}
func (m *mockIneffRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error) {
	return m.ineffs, nil
}
func (m *mockIneffRepo) ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error) {
	return nil, nil
}
func (m *mockIneffRepo) Resolve(ctx context.Context, id string, resolvedAt time.Time) error {
	return nil
}
func (m *mockIneffRepo) CountActive(ctx context.Context, clusterID string) (int, error) {
	return len(m.ineffs), nil
}

type mockCache struct {
	data map[string]interface{}
}

func newMockCache() *mockCache                                                   { return &mockCache{data: make(map[string]interface{})} }
func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}
func (m *mockCache) Delete(ctx context.Context, key string) error         { return nil }
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }

func defaultDetectionConfig() config.DetectionConfig {
	return config.DetectionConfig{
		EvaluationInterval:         60 * time.Second,
		GPUUtilizationThreshold:    20.0,
		MemoryUtilizationThreshold: 15.0,
		IdleTimeThreshold:          30 * time.Minute,
		RetryThreshold:             3,
	}
}

// ---- Tests ----

func TestDetectUnderutilizedGPU(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 10.0, MemoryUtilization: 50.0, Temperature: 40.0, Timestamp: time.Now()},
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 12.0, MemoryUtilization: 50.0, Temperature: 42.0, Timestamp: time.Now()},
			{GPUID: "gpu-002", ClusterID: "c1", NodeID: "n1", Utilization: 85.0, MemoryUtilization: 70.0, Temperature: 72.0, Timestamp: time.Now()},
		},
	}
	ineffRepo := &mockIneffRepo{}
	svc := NewService(repo, ineffRepo, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	// gpu-001 should be flagged as underutilized (avg 11%), gpu-002 should not
	found := false
	for _, r := range results {
		if r.Type == detection.TypeUnderutilizedGPU && r.GPUID == "gpu-001" {
			found = true
			if r.Severity != detection.SeverityMedium {
				t.Errorf("expected severity medium, got %s", r.Severity)
			}
		}
		if r.Type == detection.TypeUnderutilizedGPU && r.GPUID == "gpu-002" {
			t.Error("gpu-002 should NOT be flagged as underutilized")
		}
	}
	if !found {
		t.Error("expected gpu-001 to be flagged as underutilized")
	}
}

func TestDetectIdleGPU(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-idle", ClusterID: "c1", NodeID: "n1", Utilization: 0.0, RunningProcesses: 0, MemoryUtilization: 1.0, Temperature: 30.0, Timestamp: time.Now()},
			{GPUID: "gpu-idle", ClusterID: "c1", NodeID: "n1", Utilization: 0.5, RunningProcesses: 0, MemoryUtilization: 1.0, Temperature: 30.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == detection.TypeIdleGPU && r.GPUID == "gpu-idle" {
			found = true
		}
	}
	if !found {
		t.Error("expected gpu-idle to be flagged as idle")
	}
}

func TestDetectCPUStarvation(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "node-hot", Utilization: 30.0, MemoryUtilization: 60.0, Temperature: 50.0, Timestamp: time.Now()},
		},
		nodeMetrics: []metrics.NodeMetric{
			{NodeID: "node-hot", ClusterID: "c1", CPUUsage: 95.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == detection.TypeCPUStarvation && r.NodeID == "node-hot" {
			found = true
		}
	}
	if !found {
		t.Error("expected node-hot to be flagged for CPU starvation")
	}
}

func TestDetectStalledTraining(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-stall", ClusterID: "c1", NodeID: "n1", Utilization: 2.0, MemoryUtilization: 65.0, RunningProcesses: 1, JobID: "job-1", Temperature: 45.0, Timestamp: time.Now()},
			{GPUID: "gpu-stall", ClusterID: "c1", NodeID: "n1", Utilization: 3.0, MemoryUtilization: 60.0, RunningProcesses: 1, JobID: "job-1", Temperature: 44.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == detection.TypeStalledTraining && r.GPUID == "gpu-stall" {
			found = true
			if r.Severity != detection.SeverityCritical {
				t.Errorf("stalled training should be critical, got %s", r.Severity)
			}
		}
	}
	if !found {
		t.Error("expected gpu-stall to be flagged as stalled training")
	}
}

func TestDetectThermalSpike(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-hot", ClusterID: "c1", NodeID: "n1", Utilization: 95.0, MemoryUtilization: 80.0, Temperature: 95.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	found := false
	for _, r := range results {
		if r.Type == detection.TypeAbnormalSpike && r.GPUID == "gpu-hot" {
			found = true
		}
	}
	if !found {
		t.Error("expected gpu-hot to be flagged for thermal spike")
	}
}

func TestDeduplication(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 5.0, MemoryUtilization: 50.0, Temperature: 40.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	// First evaluation should detect
	results1, _ := svc.Evaluate(context.Background())

	// Second evaluation with same data should deduplicate
	results2, _ := svc.Evaluate(context.Background())

	if len(results1) == 0 {
		t.Fatal("first evaluation should detect inefficiencies")
	}
	if len(results2) != 0 {
		t.Errorf("second evaluation should deduplicate, got %d new results", len(results2))
	}
}

func TestNoMetrics(t *testing.T) {
	repo := &mockMetricsRepo{gpuMetrics: nil}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for no metrics, got %d", len(results))
	}
}

func TestHealthyCluster(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 85.0, MemoryUtilization: 70.0, Temperature: 72.0, RunningProcesses: 2, Timestamp: time.Now()},
			{GPUID: "gpu-002", ClusterID: "c1", NodeID: "n1", Utilization: 90.0, MemoryUtilization: 75.0, Temperature: 70.0, RunningProcesses: 1, Timestamp: time.Now()},
		},
		nodeMetrics: []metrics.NodeMetric{
			{NodeID: "n1", ClusterID: "c1", CPUUsage: 45.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	results, err := svc.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("healthy cluster should produce no inefficiencies, got %d", len(results))
	}
}

func TestOnDetectionListener(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 0.0, RunningProcesses: 0, MemoryUtilization: 1.0, Temperature: 30.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	var received []detection.Inefficiency
	svc.OnDetection(func(ctx context.Context, ineff detection.Inefficiency) {
		received = append(received, ineff)
	})

	svc.Evaluate(context.Background())

	if len(received) == 0 {
		t.Error("listener should have been called with detected inefficiencies")
	}
}

func TestStats(t *testing.T) {
	repo := &mockMetricsRepo{
		gpuMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-001", ClusterID: "c1", NodeID: "n1", Utilization: 5.0, MemoryUtilization: 50.0, Temperature: 40.0, Timestamp: time.Now()},
		},
	}
	svc := NewService(repo, &mockIneffRepo{}, newMockCache(), logger.New("development"), defaultDetectionConfig())

	svc.Evaluate(context.Background())

	stats := svc.Stats()
	if stats["total_evaluations"].(int64) != 1 {
		t.Errorf("total_evaluations = %v, want 1", stats["total_evaluations"])
	}
	if stats["rules_count"].(int) != 6 {
		t.Errorf("rules_count = %v, want 6", stats["rules_count"])
	}
}
