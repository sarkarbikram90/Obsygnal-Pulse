package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	recsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/recommendation"
)

type mockRecRepo struct {
	inserted []recommendation.Recommendation
}

func (m *mockRecRepo) InsertBatch(ctx context.Context, recs []recommendation.Recommendation) error {
	m.inserted = append(m.inserted, recs...)
	return nil
}
func (m *mockRecRepo) Insert(ctx context.Context, rec *recommendation.Recommendation) error {
	return nil
}
func (m *mockRecRepo) GetByID(ctx context.Context, id string) (*recommendation.Recommendation, error) {
	return nil, nil
}
func (m *mockRecRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	return nil, nil
}
func (m *mockRecRepo) Dismiss(ctx context.Context, id, dismissedBy, reason string) error { return nil }
func (m *mockRecRepo) Apply(ctx context.Context, id string) error                        { return nil }
func (m *mockRecRepo) CountActive(ctx context.Context, clusterID string) (int, error)    { return 0, nil }

func TestRecommendationGeneration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping recommendation test. Set INTEGRATION_TEST=1 to run.")
	}

	log := logger.New("development")
	recRepo := &mockRecRepo{}
	cache := &mockCache{} // reusing mockCache from detection_test.go

	svc := recsvc.NewService(recRepo, cache, log)

	ineff := detection.Inefficiency{
		ID:             uuid.New().String(),
		Type:           detection.TypeIdleGPU,
		Severity:       detection.SeverityHigh,
		ClusterID:      "test-cluster",
		NodeID:         "node-1",
		GPUID:          "gpu-1",
		DetectedAt:     time.Now(),
		CurrentValue:   1.5,
		ThresholdValue: 2.0,
	}

	recs, err := svc.GenerateFromInefficiency(context.Background(), ineff)
	if err != nil {
		t.Fatalf("Failed to generate recommendation: %v", err)
	}

	if len(recs) == 0 {
		t.Fatalf("Expected at least one recommendation, got 0")
	}

	rec := recs[0]
	if rec.Category != recommendation.CategoryReduceIdleTime {
		t.Errorf("Expected category %s, got %s", recommendation.CategoryReduceIdleTime, rec.Category)
	}

	if rec.EstimatedSavings.CostPerHourUSD <= 0 {
		t.Errorf("Expected positive cost estimate, got %f", rec.EstimatedSavings.CostPerHourUSD)
	}
}
