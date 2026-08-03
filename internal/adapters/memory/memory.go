// Package memory implements thread-safe stateful in-memory repositories.
// This allows Obsygnal-Pulse to operate with 100% data consistency standalone,
// in automated testing, or when ClickHouse/Postgres clusters are starting up.
package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// Repository Store manages in-memory data tables.
type RepositoryStore struct {
	gpuMetrics   []metrics.GPUMetric
	nodeMetrics  []metrics.NodeMetric
	ineffs       map[string]*detection.Inefficiency
	recs         map[string]*recommendation.Recommendation
	alerts       []ports.Alert
	mu           sync.RWMutex
	log          *logger.Logger
}

// NewRepositoryStore creates a initialized memory store.
func NewRepositoryStore(log *logger.Logger) *RepositoryStore {
	store := &RepositoryStore{
		ineffs: make(map[string]*detection.Inefficiency),
		recs:   make(map[string]*recommendation.Recommendation),
		log:    log.WithComponent("memory-repo"),
	}
	store.seedInitialData()
	return store
}

func (s *RepositoryStore) seedInitialData() {
	now := time.Now()
	clusterID := "sim-cluster-01"

	// Seed Sample Inefficiencies
	ineff1 := detection.Inefficiency{
		ID:              "ineff-001",
		Type:            detection.TypeIdleGPU,
		Severity:        detection.SeverityHigh,
		ClusterID:       clusterID,
		NodeID:          "gpu-node-01",
		GPUID:           "gpu-node-01-gpu-3",
		Description:     "GPU gpu-node-01-gpu-3 is idle (0% utilization, 0 running processes)",
		AffectedMetrics: []string{"gpu_utilization", "running_processes"},
		CurrentValue:    0.0,
		ThresholdValue:  2.0,
		DetectedAt:      now.Add(-25 * time.Minute),
		IsActive:        true,
	}
	s.ineffs[ineff1.ID] = &ineff1

	ineff2 := detection.Inefficiency{
		ID:              "ineff-002",
		Type:            detection.TypeUnderutilizedGPU,
		Severity:        detection.SeverityMedium,
		ClusterID:       clusterID,
		NodeID:          "gpu-node-02",
		GPUID:           "gpu-node-02-gpu-1",
		JobID:           "job-llama3-finetune-891",
		Description:     "GPU gpu-node-02-gpu-1 average utilization is 12.4% (threshold: 30.0%)",
		AffectedMetrics: []string{"gpu_utilization"},
		CurrentValue:    12.4,
		ThresholdValue:  30.0,
		DetectedAt:      now.Add(-14 * time.Minute),
		IsActive:        true,
	}
	s.ineffs[ineff2.ID] = &ineff2

	// Seed Recommendations
	rec1 := recommendation.Recommendation{
		ID:             "rec-001",
		Category:       recommendation.CategoryReduceIdleTime,
		Status:         recommendation.StatusActive,
		Title:          "Release Idle H100 SXM GPU",
		Description:    "GPU gpu-node-01-gpu-3 has been idle for 25 minutes. Terminate pod allocation to avoid $5.12/hr waste.",
		Explanation:    "Continuous 0% compute utilization detected while instance cost accumulates. Releasing allocation saves $3,737.60/month.",
		ConfidenceScore: 0.96,
		EstimatedSavings: recommendation.EstimatedSavings{
			CostPerHourUSD:   5.12,
			CostPerMonthUSD:  3737.60,
			GPUHoursPerMonth: 730,
		},
		AffectedMetrics: []string{"gpu_utilization", "cost_usd"},
		ClusterID:       clusterID,
		NodeID:          "gpu-node-01",
		GPUID:           "gpu-node-01-gpu-3",
		InefficiencyID:  ineff1.ID,
		CreatedAt:       now.Add(-20 * time.Minute),
		ExpiresAt:       now.Add(24 * time.Hour),
	}
	s.recs[rec1.ID] = &rec1

	rec2 := recommendation.Recommendation{
		ID:             "rec-002",
		Category:       recommendation.CategoryIncreaseBatchSize,
		Status:         recommendation.StatusActive,
		Title:          "Increase PyTorch Micro-Batch Size",
		Description:    "Increase micro-batch size from 16 to 64 on job-llama3-finetune-891 to saturate GPU tensor cores.",
		Explanation:    "GPU memory utilization is only 22.1% out of 80GB VRAM. Doubling batch size increases GPU throughput by 3.2x.",
		ConfidenceScore: 0.88,
		EstimatedSavings: recommendation.EstimatedSavings{
			CostPerHourUSD:   2.40,
			CostPerMonthUSD:  1752.00,
			GPUHoursPerMonth: 210,
		},
		AffectedMetrics: []string{"gpu_utilization", "gpu_memory_utilization"},
		ClusterID:       clusterID,
		NodeID:          "gpu-node-02",
		GPUID:           "gpu-node-02-gpu-1",
		JobID:           "job-llama3-finetune-891",
		InefficiencyID:  ineff2.ID,
		CreatedAt:       now.Add(-10 * time.Minute),
		ExpiresAt:       now.Add(24 * time.Hour),
	}
	s.recs[rec2.ID] = &rec2

	// Seed Alerts
	s.alerts = append(s.alerts, ports.Alert{
		ID:        "alert-101",
		Type:      "idle_gpu",
		Severity:  "critical",
		Title:     "Idle GPU Alert: gpu-node-01-gpu-3",
		Message:   "GPU utilization dropped to 0% for > 20 minutes while costing $5.12/hr.",
		ClusterID: clusterID,
		NodeID:    "gpu-node-01",
		GPUID:     "gpu-node-01-gpu-3",
		FiredAt:   now.Add(-20 * time.Minute),
		Channels:  []string{"slack", "email"},
	})
}

// ============================================================
// METRICS REPOSITORY
// ============================================================

func (s *RepositoryStore) InsertGPUMetrics(ctx context.Context, m []metrics.GPUMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gpuMetrics = append(s.gpuMetrics, m...)
	return nil
}

func (s *RepositoryStore) InsertNodeMetrics(ctx context.Context, m []metrics.NodeMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeMetrics = append(s.nodeMetrics, m...)
	return nil
}

func (s *RepositoryStore) QueryGPUMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.GPUMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gpuMetrics, nil
}

func (s *RepositoryStore) QueryNodeMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.NodeMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodeMetrics, nil
}

func (s *RepositoryStore) GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var total float64
	var count int
	for _, m := range s.gpuMetrics {
		if m.GPUID == gpuID && m.Timestamp.After(from) && m.Timestamp.Before(to) {
			total += m.Utilization
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / float64(count), nil
}

func (s *RepositoryStore) GetClusterOverview(ctx context.Context, clusterID string) (*ports.ClusterOverview, error) {
	return &ports.ClusterOverview{
		ClusterID:         clusterID,
		TotalNodes:        4,
		HealthyNodes:      4,
		TotalGPUs:         32,
		ActiveGPUs:        24,
		IdleGPUs:          8,
		AvgUtilization:    44.8,
		AvgMemoryUsage:    58.2,
		ActiveJobs:        6,
		FailedJobs24h:     0,
		InefficiencyScore: 55.2,
		EstimatedWaste:    6820.00,
		Timestamp:         time.Now(),
	}, nil
}

// ============================================================
// INEFFICIENCY REPOSITORY
// ============================================================

func (s *RepositoryStore) Insert(ctx context.Context, ineff *detection.Inefficiency) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ineff.ID == "" {
		ineff.ID = uuid.New().String()
	}
	s.ineffs[ineff.ID] = ineff
	return nil
}

func (s *RepositoryStore) InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error {
	for i := range ineffs {
		_ = s.Insert(ctx, &ineffs[i])
	}
	return nil
}

func (s *RepositoryStore) GetByID(ctx context.Context, id string) (*detection.Inefficiency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ineff, exists := s.ineffs[id]
	if !exists {
		return nil, fmt.Errorf("inefficiency %s not found", id)
	}
	return ineff, nil
}

func (s *RepositoryStore) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []detection.Inefficiency
	for _, ineff := range s.ineffs {
		if ineff.IsActive && (clusterID == "" || ineff.ClusterID == clusterID) {
			list = append(list, *ineff)
		}
	}
	return list, nil
}

func (s *RepositoryStore) ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []detection.Inefficiency
	for _, ineff := range s.ineffs {
		if ineff.Type == itype && (clusterID == "" || ineff.ClusterID == clusterID) {
			list = append(list, *ineff)
		}
	}
	return list, nil
}

func (s *RepositoryStore) Resolve(ctx context.Context, id string, resolvedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ineff, ok := s.ineffs[id]; ok {
		ineff.IsActive = false
		ineff.ResolvedAt = &resolvedAt
	}
	return nil
}

func (s *RepositoryStore) CountActive(ctx context.Context, clusterID string) (int, error) {
	list, err := s.ListActive(ctx, clusterID, 1000, 0)
	return len(list), err
}

// ============================================================
// RECOMMENDATION REPOSITORY
// ============================================================

func (s *RepositoryStore) InsertRec(ctx context.Context, rec *recommendation.Recommendation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	s.recs[rec.ID] = rec
	return nil
}

func (s *RepositoryStore) InsertRecBatch(ctx context.Context, recs []recommendation.Recommendation) error {
	for i := range recs {
		_ = s.InsertRec(ctx, &recs[i])
	}
	return nil
}

func (s *RepositoryStore) ListActiveRecs(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []recommendation.Recommendation
	for _, rec := range s.recs {
		if rec.Status == recommendation.StatusActive && (clusterID == "" || rec.ClusterID == clusterID) {
			list = append(list, *rec)
		}
	}
	return list, nil
}

func (s *RepositoryStore) DismissRec(ctx context.Context, id, dismissedBy, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec, ok := s.recs[id]; ok {
		rec.Status = recommendation.StatusDismissed
	}
	return nil
}

func (s *RepositoryStore) ApplyRec(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec, ok := s.recs[id]; ok {
		rec.Status = recommendation.StatusApplied
	}
	return nil
}

func (s *RepositoryStore) CountActiveRecs(ctx context.Context, clusterID string) (int, error) {
	list, err := s.ListActiveRecs(ctx, clusterID, 1000, 0)
	return len(list), err
}

// ============================================================
// ALERT REPOSITORY
// ============================================================

func (s *RepositoryStore) InsertAlert(ctx context.Context, alert *ports.Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	s.alerts = append([]ports.Alert{*alert}, s.alerts...)
	return nil
}

func (s *RepositoryStore) ListRecentAlerts(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alerts, nil
}

func (s *RepositoryStore) AcknowledgeAlert(ctx context.Context, id, acknowledgedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for i := range s.alerts {
		if s.alerts[i].ID == id {
			s.alerts[i].AcknowledgedAt = &now
			s.alerts[i].AcknowledgedBy = acknowledgedBy
		}
	}
	return nil
}
