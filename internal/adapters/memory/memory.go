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
	domainauth "github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// RepositoryStore manages in-memory data tables.
type RepositoryStore struct {
	gpuMetrics  []metrics.GPUMetric
	nodeMetrics []metrics.NodeMetric
	ineffs      map[string]*detection.Inefficiency
	recs        map[string]*recommendation.Recommendation
	alerts      []ports.Alert
	users       map[string]*domainauth.User // key: userID or email
	orgs        map[string]*domainauth.Organization
	orgMembers  map[string]*domainauth.OrgMember // key: orgID:userID
	apiKeys     map[string]*domainauth.APIKey    // key: keyID or keyHash
	auditEvents []domainauth.AuditEvent
	usage       *ports.UsageSummary
	mu          sync.RWMutex
	log         *logger.Logger
}

// InefficiencyMemoryRepo adapts RepositoryStore to ports.InefficiencyRepository.
type InefficiencyMemoryRepo struct {
	store *RepositoryStore
}

// RecommendationMemoryRepo adapts RepositoryStore to ports.RecommendationRepository.
type RecommendationMemoryRepo struct {
	store *RepositoryStore
}

// AlertMemoryRepo adapts RepositoryStore to ports.AlertRepository.
type AlertMemoryRepo struct {
	store *RepositoryStore
}

// AuthMemoryRepo adapts RepositoryStore to ports.AuthRepository.
type AuthMemoryRepo struct {
	store *RepositoryStore
}

// APIKeyMemoryRepo adapts RepositoryStore to ports.APIKeyRepository.
type APIKeyMemoryRepo struct {
	store *RepositoryStore
}

// AuditLogMemoryRepo adapts RepositoryStore to ports.AuditLogRepository.
type AuditLogMemoryRepo struct {
	store *RepositoryStore
}

// UsageMemoryRepo adapts RepositoryStore to ports.UsageRepository.
type UsageMemoryRepo struct {
	store *RepositoryStore
}

// NewRepositoryStore creates an initialized memory store.
func NewRepositoryStore(log *logger.Logger) *RepositoryStore {
	store := &RepositoryStore{
		ineffs:     make(map[string]*detection.Inefficiency),
		recs:       make(map[string]*recommendation.Recommendation),
		users:      make(map[string]*domainauth.User),
		orgs:       make(map[string]*domainauth.Organization),
		orgMembers: make(map[string]*domainauth.OrgMember),
		apiKeys:    make(map[string]*domainauth.APIKey),
		log:        log.WithComponent("memory-repo"),
	}
	store.seedInitialData()
	return store
}

func (s *RepositoryStore) InefficiencyRepo() ports.InefficiencyRepository {
	return &InefficiencyMemoryRepo{store: s}
}

func (s *RepositoryStore) RecommendationRepo() ports.RecommendationRepository {
	return &RecommendationMemoryRepo{store: s}
}

func (s *RepositoryStore) AlertRepo() ports.AlertRepository {
	return &AlertMemoryRepo{store: s}
}

func (s *RepositoryStore) AuthRepo() ports.AuthRepository {
	return &AuthMemoryRepo{store: s}
}

func (s *RepositoryStore) APIKeyRepo() ports.APIKeyRepository {
	return &APIKeyMemoryRepo{store: s}
}

func (s *RepositoryStore) AuditLogRepo() ports.AuditLogRepository {
	return &AuditLogMemoryRepo{store: s}
}

func (s *RepositoryStore) UsageRepo() ports.UsageRepository {
	return &UsageMemoryRepo{store: s}
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
		ID:              "rec-001",
		Category:        recommendation.CategoryReduceIdleTime,
		Status:          recommendation.StatusActive,
		Title:           "Release Idle H100 SXM GPU",
		Description:     "GPU gpu-node-01-gpu-3 has been idle for 25 minutes. Terminate pod allocation to avoid $5.12/hr waste.",
		Explanation:     "Continuous 0% compute utilization detected while instance cost accumulates. Releasing allocation saves $3,737.60/month.",
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
		ID:              "rec-002",
		Category:        recommendation.CategoryIncreaseBatchSize,
		Status:          recommendation.StatusActive,
		Title:           "Increase PyTorch Micro-Batch Size",
		Description:     "Increase micro-batch size from 16 to 64 on job-llama3-finetune-891 to saturate GPU tensor cores.",
		Explanation:     "GPU memory utilization is only 22.1% out of 80GB VRAM. Doubling batch size increases GPU throughput by 3.2x.",
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
// METRICS REPOSITORY (directly on RepositoryStore)
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

func (r *InefficiencyMemoryRepo) Insert(ctx context.Context, ineff *detection.Inefficiency) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if ineff.ID == "" {
		ineff.ID = uuid.New().String()
	}
	r.store.ineffs[ineff.ID] = ineff
	return nil
}

func (r *InefficiencyMemoryRepo) InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error {
	for i := range ineffs {
		_ = r.Insert(ctx, &ineffs[i])
	}
	return nil
}

func (r *InefficiencyMemoryRepo) GetByID(ctx context.Context, id string) (*detection.Inefficiency, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	ineff, exists := r.store.ineffs[id]
	if !exists {
		return nil, fmt.Errorf("inefficiency %s not found", id)
	}
	return ineff, nil
}

func (r *InefficiencyMemoryRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var list []detection.Inefficiency
	for _, ineff := range r.store.ineffs {
		if ineff.IsActive && (clusterID == "" || ineff.ClusterID == clusterID) {
			list = append(list, *ineff)
		}
	}
	return list, nil
}

func (r *InefficiencyMemoryRepo) ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var list []detection.Inefficiency
	for _, ineff := range r.store.ineffs {
		if ineff.Type == itype && (clusterID == "" || ineff.ClusterID == clusterID) {
			list = append(list, *ineff)
		}
	}
	return list, nil
}

func (r *InefficiencyMemoryRepo) Resolve(ctx context.Context, id string, resolvedAt time.Time) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if ineff, ok := r.store.ineffs[id]; ok {
		ineff.IsActive = false
		ineff.ResolvedAt = &resolvedAt
	}
	return nil
}

func (r *InefficiencyMemoryRepo) CountActive(ctx context.Context, clusterID string) (int, error) {
	list, err := r.ListActive(ctx, clusterID, 1000, 0)
	return len(list), err
}

// ============================================================
// RECOMMENDATION REPOSITORY
// ============================================================

func (r *RecommendationMemoryRepo) Insert(ctx context.Context, rec *recommendation.Recommendation) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	r.store.recs[rec.ID] = rec
	return nil
}

func (r *RecommendationMemoryRepo) InsertBatch(ctx context.Context, recs []recommendation.Recommendation) error {
	for i := range recs {
		_ = r.Insert(ctx, &recs[i])
	}
	return nil
}

func (r *RecommendationMemoryRepo) GetByID(ctx context.Context, id string) (*recommendation.Recommendation, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	rec, exists := r.store.recs[id]
	if !exists {
		return nil, fmt.Errorf("recommendation %s not found", id)
	}
	return rec, nil
}

func (r *RecommendationMemoryRepo) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var list []recommendation.Recommendation
	for _, rec := range r.store.recs {
		if rec.Status == recommendation.StatusActive && (clusterID == "" || rec.ClusterID == clusterID) {
			list = append(list, *rec)
		}
	}
	return list, nil
}

func (r *RecommendationMemoryRepo) Dismiss(ctx context.Context, id, dismissedBy, reason string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if rec, ok := r.store.recs[id]; ok {
		rec.Status = recommendation.StatusDismissed
	}
	return nil
}

func (r *RecommendationMemoryRepo) Apply(ctx context.Context, id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if rec, ok := r.store.recs[id]; ok {
		rec.Status = recommendation.StatusApplied
	}
	return nil
}

func (r *RecommendationMemoryRepo) CountActive(ctx context.Context, clusterID string) (int, error) {
	list, err := r.ListActive(ctx, clusterID, 1000, 0)
	return len(list), err
}

// ============================================================
// ALERT REPOSITORY
// ============================================================

func (r *AlertMemoryRepo) Insert(ctx context.Context, alert *ports.Alert) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	r.store.alerts = append([]ports.Alert{*alert}, r.store.alerts...)
	return nil
}

func (r *AlertMemoryRepo) ListRecent(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.alerts, nil
}

func (r *AlertMemoryRepo) Acknowledge(ctx context.Context, id, acknowledgedBy string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	now := time.Now()
	for i := range r.store.alerts {
		if r.store.alerts[i].ID == id {
			r.store.alerts[i].AcknowledgedAt = &now
			r.store.alerts[i].AcknowledgedBy = acknowledgedBy
		}
	}
	return nil
}

// ============================================================
// AUTH REPOSITORY
// ============================================================

func (r *AuthMemoryRepo) CreateUser(ctx context.Context, user *domainauth.User) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	r.store.users[user.ID] = user
	r.store.users["email:"+user.Email] = user
	return nil
}

func (r *AuthMemoryRepo) GetUserByEmail(ctx context.Context, email string) (*domainauth.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	user, exists := r.store.users["email:"+email]
	if !exists {
		return nil, fmt.Errorf("user with email %s not found", email)
	}
	return user, nil
}

func (r *AuthMemoryRepo) GetUserByID(ctx context.Context, id string) (*domainauth.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	user, exists := r.store.users[id]
	if !exists {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return user, nil
}

func (r *AuthMemoryRepo) CreateOrg(ctx context.Context, org *domainauth.Organization) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if org.ID == "" {
		org.ID = uuid.New().String()
	}
	r.store.orgs[org.ID] = org
	return nil
}

func (r *AuthMemoryRepo) GetOrgByID(ctx context.Context, id string) (*domainauth.Organization, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	org, exists := r.store.orgs[id]
	if !exists {
		return nil, fmt.Errorf("organization %s not found", id)
	}
	return org, nil
}

func (r *AuthMemoryRepo) AddOrgMember(ctx context.Context, member *domainauth.OrgMember) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	key := fmt.Sprintf("%s:%s", member.OrgID, member.UserID)
	r.store.orgMembers[key] = member
	return nil
}

func (r *AuthMemoryRepo) GetOrgMember(ctx context.Context, orgID, userID string) (*domainauth.OrgMember, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", orgID, userID)
	member, exists := r.store.orgMembers[key]
	if !exists {
		return nil, fmt.Errorf("org member not found for org %s and user %s", orgID, userID)
	}
	return member, nil
}

// ============================================================
// API KEY REPOSITORY
// ============================================================

func (r *APIKeyMemoryRepo) CreateAPIKey(ctx context.Context, key *domainauth.APIKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	r.store.apiKeys[key.ID] = key
	if key.KeyHash != "" {
		r.store.apiKeys["hash:"+key.KeyHash] = key
	}
	return nil
}

func (r *APIKeyMemoryRepo) ListAPIKeys(ctx context.Context, orgID string) ([]domainauth.APIKey, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var keys []domainauth.APIKey
	seen := make(map[string]bool)
	for id, k := range r.store.apiKeys {
		if !seen[k.ID] && k.OrgID == orgID && id == k.ID {
			seen[k.ID] = true
			keys = append(keys, *k)
		}
	}
	return keys, nil
}

func (r *APIKeyMemoryRepo) RevokeAPIKey(ctx context.Context, id, orgID string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if k, exists := r.store.apiKeys[id]; exists {
		delete(r.store.apiKeys, id)
		delete(r.store.apiKeys, "hash:"+k.KeyHash)
	}
	return nil
}

func (r *APIKeyMemoryRepo) GetAPIKeyByHash(ctx context.Context, hash string) (*domainauth.APIKey, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	key, exists := r.store.apiKeys["hash:"+hash]
	if !exists {
		return nil, fmt.Errorf("API key hash not found")
	}
	return key, nil
}

// ============================================================
// AUDIT LOG REPOSITORY
// ============================================================

func (r *AuditLogMemoryRepo) InsertAuditEvent(ctx context.Context, event *domainauth.AuditEvent) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	r.store.auditEvents = append([]domainauth.AuditEvent{*event}, r.store.auditEvents...)
	return nil
}

func (r *AuditLogMemoryRepo) ListAuditEvents(ctx context.Context, orgID string, limit int) ([]domainauth.AuditEvent, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var result []domainauth.AuditEvent
	for _, evt := range r.store.auditEvents {
		if orgID == "" || evt.OrgID == orgID {
			result = append(result, evt)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// ============================================================
// USAGE REPOSITORY
// ============================================================

func (r *UsageMemoryRepo) GetUsageSummary(ctx context.Context, orgID string) (*ports.UsageSummary, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	if r.store.usage != nil && (r.store.usage.OrganizationID == orgID || orgID == "") {
		return r.store.usage, nil
	}
	return &ports.UsageSummary{
		OrganizationID:      orgID,
		Tier:                "pro",
		BillingCycle:        time.Now().Format("January 2006"),
		MonitoredGPUs:       16,
		GPUHoursUsed:        3420.5,
		GPUHoursLimit:       5000.0,
		EstimatedCostUSD:    12450.00,
		EstimatedSavingsUSD: 3820.00,
		ActiveClusters:      2,
		UpdatedAt:           time.Now(),
	}, nil
}

func (r *UsageMemoryRepo) RecordUsage(ctx context.Context, orgID string, gpuHours float64) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.usage == nil {
		r.store.usage = &ports.UsageSummary{
			OrganizationID: orgID,
			Tier:           "pro",
			BillingCycle:   time.Now().Format("January 2006"),
			GPUHoursUsed:   gpuHours,
			UpdatedAt:      time.Now(),
		}
	} else {
		r.store.usage.GPUHoursUsed += gpuHours
		r.store.usage.UpdatedAt = time.Now()
	}
	return nil
}
