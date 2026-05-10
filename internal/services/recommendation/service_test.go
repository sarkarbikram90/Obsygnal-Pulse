package recommendation

import (
	"context"
	"testing"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/detection"
	"github.com/obsygnal/obsygnal-pulse/internal/domain/recommendation"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
)

// ---- Mocks ----

type mockRecRepo struct {
	recs []recommendation.Recommendation
}

func (m *mockRecRepo) Insert(ctx context.Context, rec *recommendation.Recommendation) error {
	return nil
}
func (m *mockRecRepo) InsertBatch(ctx context.Context, recs []recommendation.Recommendation) error {
	m.recs = append(m.recs, recs...)
	return nil
}
func (m *mockRecRepo) GetByID(ctx context.Context, id string) (*recommendation.Recommendation, error) {
	return nil, nil
}
func (m *mockRecRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	return m.recs, nil
}
func (m *mockRecRepo) Dismiss(ctx context.Context, id, dismissedBy, reason string) error { return nil }
func (m *mockRecRepo) Apply(ctx context.Context, id string) error                        { return nil }
func (m *mockRecRepo) CountActive(ctx context.Context, clusterID string) (int, error) {
	return len(m.recs), nil
}

type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockCache) Delete(ctx context.Context, key string) error         { return nil }
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }

// ---- Tests ----

func TestGenerateFromUnderutilizedGPU(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:             "ineff-001",
		Type:           detection.TypeUnderutilizedGPU,
		Severity:       detection.SeverityMedium,
		ClusterID:      "c1",
		NodeID:         "n1",
		GPUID:          "gpu-001",
		CurrentValue:   10.0,
		ThresholdValue: 20.0,
		DetectedAt:     time.Now(),
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("GenerateFromInefficiency() error = %v", err)
	}

	// Should produce 2 recommendations: reduce allocation + increase batch size
	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}

	categories := make(map[recommendation.Category]bool)
	for _, r := range recs {
		categories[r.Category] = true

		if r.ConfidenceScore <= 0 || r.ConfidenceScore > 1.0 {
			t.Errorf("confidence score out of range: %f", r.ConfidenceScore)
		}
		if r.ClusterID != "c1" {
			t.Errorf("expected cluster_id c1, got %s", r.ClusterID)
		}
		if r.InefficiencyID != "ineff-001" {
			t.Errorf("expected inefficiency_id ineff-001, got %s", r.InefficiencyID)
		}
		if r.Status != recommendation.StatusActive {
			t.Errorf("expected status active, got %s", r.Status)
		}
	}

	if !categories[recommendation.CategoryReduceAllocation] {
		t.Error("expected reduce_gpu_allocation recommendation")
	}
	if !categories[recommendation.CategoryIncreaseBatchSize] {
		t.Error("expected increase_batch_size recommendation")
	}
}

func TestGenerateFromIdleGPU(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:           "ineff-002",
		Type:         detection.TypeIdleGPU,
		Severity:     detection.SeverityHigh,
		ClusterID:    "c1",
		NodeID:       "n1",
		GPUID:        "gpu-idle",
		CurrentValue: 0.0,
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}

	if recs[0].Category != recommendation.CategoryReduceIdleTime {
		t.Errorf("expected reduce_idle_runtime, got %s", recs[0].Category)
	}

	if recs[0].ConfidenceScore < 0.9 {
		t.Errorf("idle GPU recommendation should have high confidence, got %f", recs[0].ConfidenceScore)
	}

	if recs[0].EstimatedSavings.CostPerMonthUSD <= 0 {
		t.Error("idle GPU should have non-zero monthly cost savings")
	}
}

func TestGenerateFromCPUStarvation(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:             "ineff-003",
		Type:           detection.TypeCPUStarvation,
		ClusterID:      "c1",
		NodeID:         "node-hot",
		CurrentValue:   95.0,
		ThresholdValue: 90.0,
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].Category != recommendation.CategoryInvestigateCPU {
		t.Errorf("expected investigate_cpu_bottleneck, got %s", recs[0].Category)
	}
}

func TestGenerateFromStalledTraining(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:           "ineff-004",
		Type:         detection.TypeStalledTraining,
		Severity:     detection.SeverityCritical,
		ClusterID:    "c1",
		NodeID:       "n1",
		GPUID:        "gpu-stall",
		CurrentValue: 2.0,
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].ConfidenceScore < 0.85 {
		t.Errorf("stalled training should have high confidence, got %f", recs[0].ConfidenceScore)
	}
}

func TestGenerateFromThermalSpike(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:           "ineff-005",
		Type:         detection.TypeAbnormalSpike,
		Severity:     detection.SeverityCritical,
		ClusterID:    "c1",
		GPUID:        "gpu-hot",
		CurrentValue: 95.0,
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
}

func TestNoRecommendationsForUnknownType(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		Type: detection.InefficiencyType("unknown_type"),
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if recs != nil {
		t.Errorf("expected nil recommendations for unknown type, got %d", len(recs))
	}
}

func TestConfidenceScoreRange(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	// Test with extreme values
	testCases := []detection.Inefficiency{
		{ID: "1", Type: detection.TypeUnderutilizedGPU, ClusterID: "c1", GPUID: "g1", CurrentValue: 0.1, ThresholdValue: 20.0},
		{ID: "2", Type: detection.TypeUnderutilizedGPU, ClusterID: "c1", GPUID: "g2", CurrentValue: 19.5, ThresholdValue: 20.0},
		{ID: "3", Type: detection.TypeIdleGPU, ClusterID: "c1", GPUID: "g3", CurrentValue: 0.0},
	}

	for _, ineff := range testCases {
		recs, _ := svc.GenerateFromInefficiency(context.Background(), ineff)
		for _, r := range recs {
			if r.ConfidenceScore < 0 || r.ConfidenceScore > 1.0 {
				t.Errorf("confidence out of [0,1] range: %f for inefficiency %s", r.ConfidenceScore, ineff.ID)
			}
		}
	}
}

func TestStats(t *testing.T) {
	repo := &mockRecRepo{}
	svc := NewService(repo, &mockCache{}, logger.New("development"))

	ineff := detection.Inefficiency{
		ID:        "test",
		Type:      detection.TypeIdleGPU,
		ClusterID: "c1",
		GPUID:     "g1",
	}
	svc.GenerateFromInefficiency(context.Background(), ineff)

	stats := svc.Stats()
	if stats["total_generated"] != 1 {
		t.Errorf("total_generated = %d, want 1", stats["total_generated"])
	}
}
