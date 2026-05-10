package alerting

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/config"
	"github.com/obsygnal/obsygnal-pulse/internal/domain/detection"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// ---- Mocks ----

type mockAlertRepo struct {
	mu     sync.Mutex
	alerts []ports.Alert
}

func (m *mockAlertRepo) Insert(ctx context.Context, alert *ports.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, *alert)
	return nil
}
func (m *mockAlertRepo) ListRecent(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	return m.alerts, nil
}
func (m *mockAlertRepo) Acknowledge(ctx context.Context, id, acknowledgedBy string) error {
	return nil
}

type mockNotifier struct {
	sentMu sync.Mutex
	sent   []ports.Notification
	ntype  string
	err    error
}

func (m *mockNotifier) Send(ctx context.Context, notification ports.Notification) error {
	m.sentMu.Lock()
	defer m.sentMu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, notification)
	return nil
}
func (m *mockNotifier) Type() string { return m.ntype }

func defaultAlertConfig() config.AlertingConfig {
	return config.AlertingConfig{
		DeduplicationWindow: 15 * time.Minute,
		RateLimitPerMinute:  10,
	}
}

// ---- Tests ----

func TestHandleInefficiency_CriticalSeverity(t *testing.T) {
	repo := &mockAlertRepo{}
	notifier := &mockNotifier{ntype: "slack"}
	svc := NewService(repo, logger.New("development"), defaultAlertConfig())
	svc.AddNotifier(notifier)

	ineff := detection.Inefficiency{
		Type:           detection.TypeIdleGPU,
		Severity:       detection.SeverityCritical,
		ClusterID:      "c1",
		NodeID:         "n1",
		GPUID:          "gpu-001",
		Description:    "GPU is idle",
		CurrentValue:   0.0,
		ThresholdValue: 2.0,
	}

	svc.HandleInefficiency(context.Background(), ineff)

	// Should fire and deliver
	stats := svc.Stats()
	if stats["total_fired"] != 1 {
		t.Errorf("total_fired = %d, want 1", stats["total_fired"])
	}
	if stats["total_delivered"] != 1 {
		t.Errorf("total_delivered = %d, want 1", stats["total_delivered"])
	}

	// Notifier should have received
	if len(notifier.sent) != 1 {
		t.Errorf("notifier.sent = %d, want 1", len(notifier.sent))
	}
}

func TestHandleInefficiency_LowSeverityIgnored(t *testing.T) {
	repo := &mockAlertRepo{}
	notifier := &mockNotifier{ntype: "slack"}
	svc := NewService(repo, logger.New("development"), defaultAlertConfig())
	svc.AddNotifier(notifier)

	ineff := detection.Inefficiency{
		Type:     detection.TypeLowMemoryUsage,
		Severity: detection.SeverityLow,
	}

	svc.HandleInefficiency(context.Background(), ineff)

	// Low severity should not fire
	if svc.Stats()["total_fired"] != 0 {
		t.Error("low severity should not fire an alert")
	}
	if len(notifier.sent) != 0 {
		t.Error("low severity should not send notifications")
	}
}

func TestDeduplication(t *testing.T) {
	repo := &mockAlertRepo{}
	notifier := &mockNotifier{ntype: "slack"}
	svc := NewService(repo, logger.New("development"), defaultAlertConfig())
	svc.AddNotifier(notifier)

	ineff := detection.Inefficiency{
		Type:     detection.TypeIdleGPU,
		Severity: detection.SeverityHigh,
		GPUID:    "gpu-001",
		NodeID:   "n1",
	}

	// First alert fires
	svc.HandleInefficiency(context.Background(), ineff)
	// Second identical should be suppressed
	svc.HandleInefficiency(context.Background(), ineff)

	stats := svc.Stats()
	if stats["total_fired"] != 1 {
		t.Errorf("total_fired = %d, want 1 (dedup should suppress second)", stats["total_fired"])
	}
	if stats["total_suppressed"] != 1 {
		t.Errorf("total_suppressed = %d, want 1", stats["total_suppressed"])
	}
}

func TestRateLimiting(t *testing.T) {
	repo := &mockAlertRepo{}
	notifier := &mockNotifier{ntype: "slack"}
	cfg := defaultAlertConfig()
	cfg.RateLimitPerMinute = 2
	cfg.DeduplicationWindow = 0 // Disable dedup to test rate limiting
	svc := NewService(repo, logger.New("development"), cfg)
	svc.AddNotifier(notifier)

	// Fire 3 alerts with different keys
	for i := 0; i < 3; i++ {
		ineff := detection.Inefficiency{
			Type:     detection.TypeIdleGPU,
			Severity: detection.SeverityHigh,
			GPUID:    fmt.Sprintf("gpu-%03d", i),
			NodeID:   "n1",
		}
		svc.HandleInefficiency(context.Background(), ineff)
	}

	// 2 should deliver, 1 should be rate limited
	stats := svc.Stats()
	if stats["total_delivered"] != 2 {
		t.Errorf("total_delivered = %d, want 2", stats["total_delivered"])
	}
	if stats["total_suppressed"] != 1 {
		t.Errorf("total_suppressed = %d, want 1", stats["total_suppressed"])
	}
}

func TestMultipleNotifiers(t *testing.T) {
	repo := &mockAlertRepo{}
	slackNotifier := &mockNotifier{ntype: "slack"}
	webhookNotifier := &mockNotifier{ntype: "webhook"}
	svc := NewService(repo, logger.New("development"), defaultAlertConfig())
	svc.AddNotifier(slackNotifier)
	svc.AddNotifier(webhookNotifier)

	ineff := detection.Inefficiency{
		Type:     detection.TypeStalledTraining,
		Severity: detection.SeverityCritical,
		GPUID:    "gpu-001",
		NodeID:   "n1",
	}

	svc.HandleInefficiency(context.Background(), ineff)

	// Both notifiers should receive
	if len(slackNotifier.sent) != 1 {
		t.Errorf("slack notifications = %d, want 1", len(slackNotifier.sent))
	}
	if len(webhookNotifier.sent) != 1 {
		t.Errorf("webhook notifications = %d, want 1", len(webhookNotifier.sent))
	}

	if svc.Stats()["total_delivered"] != 2 {
		t.Errorf("total_delivered = %d, want 2 (one per channel)", svc.Stats()["total_delivered"])
	}
}

func TestAlertPersistence(t *testing.T) {
	repo := &mockAlertRepo{}
	svc := NewService(repo, logger.New("development"), defaultAlertConfig())

	ineff := detection.Inefficiency{
		Type:        detection.TypeAbnormalSpike,
		Severity:    detection.SeverityCritical,
		ClusterID:   "c1",
		NodeID:      "n1",
		GPUID:       "gpu-hot",
		Description: "Temperature exceeded threshold",
	}

	svc.HandleInefficiency(context.Background(), ineff)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.alerts) != 1 {
		t.Fatalf("expected 1 persisted alert, got %d", len(repo.alerts))
	}

	alert := repo.alerts[0]
	if alert.ClusterID != "c1" {
		t.Errorf("alert cluster_id = %s, want c1", alert.ClusterID)
	}
	if alert.GPUID != "gpu-hot" {
		t.Errorf("alert gpu_id = %s, want gpu-hot", alert.GPUID)
	}
	if alert.Severity != "critical" {
		t.Errorf("alert severity = %s, want critical", alert.Severity)
	}
}
