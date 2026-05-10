package ingestion

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/metrics"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// mockMetricsRepo is a test double for MetricsRepository.
type mockMetricsRepo struct {
	mu              sync.Mutex
	gpuMetrics      []metrics.GPUMetric
	nodeMetrics     []metrics.NodeMetric
	insertErr       error
	insertCallCount int
}

func (m *mockMetricsRepo) InsertGPUMetrics(ctx context.Context, gpuMetrics []metrics.GPUMetric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.insertCallCount++
	if m.insertErr != nil {
		return m.insertErr
	}
	m.gpuMetrics = append(m.gpuMetrics, gpuMetrics...)
	return nil
}

func (m *mockMetricsRepo) InsertNodeMetrics(ctx context.Context, nodeMetrics []metrics.NodeMetric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.insertErr != nil {
		return m.insertErr
	}
	m.nodeMetrics = append(m.nodeMetrics, nodeMetrics...)
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
	return nil, nil
}

// mockCache is a test double for CacheStore.
type mockCache struct {
	mu     sync.Mutex
	data   map[string]bool
	setErr error
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]bool)}
}

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error {
	return nil
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = true
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	return ok, nil
}

func validGPUMetric(id string) metrics.GPUMetric {
	return metrics.GPUMetric{
		ID:                id,
		ClusterID:         "test-cluster",
		NodeID:            "test-node",
		GPUID:             "gpu-" + id,
		Timestamp:         time.Now(),
		Utilization:       75.0,
		MemoryUtilization: 60.0,
		Temperature:       70.0,
	}
}

func TestProcessBatch_ValidMetrics(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)

	batch := &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{
			validGPUMetric("1"),
			validGPUMetric("2"),
			validGPUMetric("3"),
		},
		ReceivedAt: time.Now(),
		Source:     "test",
	}

	err := svc.ProcessBatch(context.Background(), batch)
	if err != nil {
		t.Fatalf("ProcessBatch() error = %v", err)
	}

	stats := svc.Stats()
	if stats["total_received"] != 3 {
		t.Errorf("total_received = %d, want 3", stats["total_received"])
	}
}

func TestProcessBatch_NilBatch(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)

	err := svc.ProcessBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("ProcessBatch(nil) should not error, got: %v", err)
	}
}

func TestProcessBatch_InvalidMetricsDropped(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)

	batch := &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{
			validGPUMetric("1"),
			{
				// Missing required fields — should be dropped
				Utilization: 50.0,
				Timestamp:   time.Now(),
			},
			validGPUMetric("3"),
		},
		ReceivedAt: time.Now(),
		Source:     "test",
	}

	err := svc.ProcessBatch(context.Background(), batch)
	if err != nil {
		t.Fatalf("ProcessBatch() error = %v", err)
	}

	stats := svc.Stats()
	if stats["total_dropped"] != 1 {
		t.Errorf("total_dropped = %d, want 1", stats["total_dropped"])
	}
}

func TestProcessBatch_DuplicateDetection(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)
	ctx := context.Background()

	metric := validGPUMetric("dup-test")

	// Process first time
	batch1 := &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{metric},
		ReceivedAt: time.Now(),
		Source:     "test",
	}
	svc.ProcessBatch(ctx, batch1)

	// Process same metric again
	batch2 := &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{metric},
		ReceivedAt: time.Now(),
		Source:     "test",
	}
	svc.ProcessBatch(ctx, batch2)

	stats := svc.Stats()
	if stats["total_duplicates"] != 1 {
		t.Errorf("total_duplicates = %d, want 1", stats["total_duplicates"])
	}
}

func TestProcessBatch_FlushOnBatchSize(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	// Set batch size to 3 to trigger flush
	svc := NewService(repo, cache, log, 3, 5*time.Second)

	batch := &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{
			validGPUMetric("1"),
			validGPUMetric("2"),
			validGPUMetric("3"),
		},
		ReceivedAt: time.Now(),
		Source:     "test",
	}

	err := svc.ProcessBatch(context.Background(), batch)
	if err != nil {
		t.Fatalf("ProcessBatch() error = %v", err)
	}

	// Should have flushed to repo
	repo.mu.Lock()
	count := len(repo.gpuMetrics)
	repo.mu.Unlock()

	if count != 3 {
		t.Errorf("flushed metrics = %d, want 3", count)
	}
}

func TestFlush_EmptyBuffer(t *testing.T) {
	repo := &mockMetricsRepo{}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)

	err := svc.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush() with empty buffer should not error, got: %v", err)
	}

	if repo.insertCallCount != 0 {
		t.Errorf("insertCallCount = %d, want 0 for empty buffer", repo.insertCallCount)
	}
}

func TestFlush_ClickHouseError(t *testing.T) {
	repo := &mockMetricsRepo{insertErr: fmt.Errorf("connection refused")}
	cache := newMockCache()
	log := logger.New("development")

	svc := NewService(repo, cache, log, 100, 5*time.Second)

	// Add metrics to buffer
	svc.ProcessBatch(context.Background(), &metrics.MetricBatch{
		GPUMetrics: []metrics.GPUMetric{validGPUMetric("1")},
		ReceivedAt: time.Now(),
		Source:     "test",
	})

	err := svc.Flush(context.Background())
	if err == nil {
		t.Error("Flush() should return error when ClickHouse fails")
	}
}
