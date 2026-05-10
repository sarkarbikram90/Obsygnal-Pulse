// Package alerting implements the alert management service.
package alerting

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	domainAlerting "github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/alerting"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// Service manages alert routing, deduplication, and delivery.
type Service struct {
	alertRepo ports.AlertRepository
	notifiers []ports.Notifier
	log       *logger.Logger
	cfg       config.AlertingConfig

	// Deduplication: track last alert time per key
	dedupMu   sync.RWMutex
	lastAlert map[string]time.Time

	// Suppression Windows
	suppressionMu sync.RWMutex
	windows       []domainAlerting.SuppressionWindow

	// Rate limiting: track alert count per channel per minute
	rateMu    sync.Mutex
	rateCount map[string]int
	rateReset time.Time

	// Stats
	totalFired      int64
	totalSuppressed int64
	totalDelivered  int64
	totalFailed     int64
}

// NewService creates a new alerting service.
func NewService(
	alertRepo ports.AlertRepository,
	log *logger.Logger,
	cfg config.AlertingConfig,
) *Service {
	return &Service{
		alertRepo: alertRepo,
		log:       log.WithComponent("alerting"),
		cfg:       cfg,
		lastAlert: make(map[string]time.Time),
		rateCount: make(map[string]int),
		rateReset: time.Now().Add(time.Minute),
	}
}

// AddNotifier registers a notification channel.
func (s *Service) AddNotifier(n ports.Notifier) {
	s.notifiers = append(s.notifiers, n)
	s.log.Info().Str("type", n.Type()).Msg("registered notification channel")
}

// AddSuppressionWindow adds a quiet period.
func (s *Service) AddSuppressionWindow(win domainAlerting.SuppressionWindow) {
	s.suppressionMu.Lock()
	defer s.suppressionMu.Unlock()
	s.windows = append(s.windows, win)
}

// HandleInefficiency processes a detected inefficiency and fires alerts if appropriate.
func (s *Service) HandleInefficiency(ctx context.Context, ineff detection.Inefficiency) {
	// Only alert on high/critical severity
	if ineff.Severity != detection.SeverityHigh && ineff.Severity != detection.SeverityCritical {
		return
	}

	// Check global suppression windows
	if s.isSuppressed() {
		s.totalSuppressed++
		s.log.Info().Msg("alert suppressed via maintenance window")
		return
	}

	// Check deduplication
	dedupKey := fmt.Sprintf("%s:%s:%s", ineff.Type, ineff.GPUID, ineff.NodeID)
	if s.isDuplicate(dedupKey) {
		s.totalSuppressed++
		s.log.Debug().Str("key", dedupKey).Msg("alert suppressed (deduplicated)")
		return
	}

	// Calculate rich payload data
	estimatedWaste := s.calculateEstimatedWaste(ineff)
	remediation := s.getRemediationGuidance(ineff)
	dashboardURL := fmt.Sprintf("http://dashboard.local/cluster/%s/gpus/%s", ineff.ClusterID, ineff.GPUID)

	// Create alert record
	alert := &ports.Alert{
		ID:        uuid.New().String(),
		Type:      string(ineff.Type),
		Severity:  string(ineff.Severity),
		Title:     fmt.Sprintf("[%s] %s", string(ineff.Severity), ineff.Description),
		Message:   ineff.Description,
		ClusterID: ineff.ClusterID,
		NodeID:    ineff.NodeID,
		GPUID:     ineff.GPUID,
		JobID:     ineff.JobID,
		FiredAt:   time.Now(),
		Labels:    ineff.Labels,
	}

	// Persist alert
	if err := s.alertRepo.Insert(ctx, alert); err != nil {
		s.log.Error().Err(err).Str("alert_id", alert.ID).Msg("failed to persist alert")
	}

	s.totalFired++

	// Dispatch to notification channels with enriched meta info
	notification := ports.Notification{
		Title:          alert.Title,
		Message:        alert.Message,
		Severity:       alert.Severity,
		EstimatedWaste: estimatedWaste,
		Remediation:    remediation,
		DashboardURL:   dashboardURL,
		Fields: map[string]string{
			"Cluster": ineff.ClusterID,
			"Node":    ineff.NodeID,
			"GPU":     ineff.GPUID,
			"Type":    string(ineff.Type),
		},
	}

	// Determine target channels based on severity (Routing logic)
	targetChannels := s.routeAlert(ineff.Severity)

	deliveredChannels := make([]string, 0)
	for _, notifier := range s.notifiers {
		// Skip if not routed for this notifier type
		if !s.sliceContains(targetChannels, notifier.Type()) {
			continue
		}

		if s.isRateLimited(notifier.Type()) {
			s.totalSuppressed++
			s.log.Warn().Str("channel", notifier.Type()).Msg("alert suppressed (rate limited)")
			continue
		}

		// Execute delivery with dynamic retry
		if err := s.sendWithRetry(ctx, notifier, notification); err != nil {
			s.totalFailed++
			s.log.Error().Err(err).Str("channel", notifier.Type()).Msg("failed to deliver alert after retries (dead-letter logged)")
		} else {
			s.totalDelivered++
			deliveredChannels = append(deliveredChannels, notifier.Type())
			s.log.Info().Str("alert_id", alert.ID).Str("channel", notifier.Type()).Msg("alert delivered")
		}
	}

	alert.Channels = deliveredChannels
	s.markAlerted(dedupKey)
}

func (s *Service) sendWithRetry(ctx context.Context, n ports.Notifier, note ports.Notification) error {
	var lastErr error
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := n.Send(ctx, note)
		if err == nil {
			return nil
		}
		lastErr = err
		s.log.Warn().Err(err).Int("attempt", attempt).Str("notifier", n.Type()).Msg("delivery retry scheduled")

		// Exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(math.Pow(2, float64(attempt))) * time.Second):
		}
	}
	return fmt.Errorf("exhausted %d attempts: %w", maxAttempts, lastErr)
}

func (s *Service) routeAlert(sev detection.Severity) []string {
	// Base route: always Slack
	channels := []string{"slack"}
	// Escalate Critical to both Slack + Webhooks / PagerDuty adapters
	if sev == detection.SeverityCritical {
		channels = append(channels, "webhook")
	}
	return channels
}

func (s *Service) calculateEstimatedWaste(ineff detection.Inefficiency) float64 {
	// Basic heuristic matching implementation for alerting payloads
	switch ineff.Type {
	case detection.TypeUnderutilizedGPU, detection.TypeIdleGPU:
		return 300.00 // Monthly waste example base
	case detection.TypeStalledTraining:
		return 500.00
	default:
		return 50.00
	}
}

func (s *Service) getRemediationGuidance(ineff detection.Inefficiency) string {
	switch ineff.Type {
	case detection.TypeUnderutilizedGPU:
		return "Downscale instance size or bundle multiple small models onto this accelerator."
	case detection.TypeIdleGPU:
		return "Deallocate unused node via control-plane CLI or terminate idle pod immediately."
	case detection.TypeStalledTraining:
		return "Investigate data loading pipelines or clear deadlock. Reboot of training pod recommended."
	default:
		return "View live GPU dashboard metrics and evaluate active workload queues."
	}
}

func (s *Service) isSuppressed() bool {
	s.suppressionMu.RLock()
	defer s.suppressionMu.RUnlock()
	now := time.Now()
	for _, win := range s.windows {
		if now.After(win.StartTime) && now.Before(win.EndTime) {
			return true
		}
	}
	return false
}

func (s *Service) isDuplicate(key string) bool {
	s.dedupMu.RLock()
	defer s.dedupMu.RUnlock()

	if last, ok := s.lastAlert[key]; ok {
		return time.Since(last) < s.cfg.DeduplicationWindow
	}
	return false
}

func (s *Service) markAlerted(key string) {
	s.dedupMu.Lock()
	defer s.dedupMu.Unlock()
	s.lastAlert[key] = time.Now()
}

func (s *Service) isRateLimited(channel string) bool {
	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := time.Now()
	if now.After(s.rateReset) {
		s.rateCount = make(map[string]int)
		s.rateReset = now.Add(time.Minute)
	}

	count := s.rateCount[channel]
	if count >= s.cfg.RateLimitPerMinute {
		return true
	}
	s.rateCount[channel]++
	return false
}

func (s *Service) sliceContains(list []string, item string) bool {
	for _, a := range list {
		if a == item {
			return true
		}
	}
	return false
}

func (s *Service) ListRecent(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	return s.alertRepo.ListRecent(ctx, clusterID, limit)
}

func (s *Service) Stats() map[string]int64 {
	return map[string]int64{
		"total_fired":      s.totalFired,
		"total_suppressed": s.totalSuppressed,
		"total_delivered":  s.totalDelivered,
		"total_failed":     s.totalFailed,
	}
}
