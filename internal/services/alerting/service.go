// Package alerting implements the alert management service.
//
// The alerting service is responsible for:
// 1. Receiving events from the detection engine
// 2. Evaluating severity against notification thresholds
// 3. Deduplicating alerts within a configurable window
// 4. Rate-limiting alerts per channel
// 5. Routing alerts to configured channels (Slack, webhooks)
// 6. Persisting alert history
//
// Design:
//   - Alerts are triggered by detection events, not polled
//   - Dedup key: type + target (GPU/node ID)
//   - Rate limiting is per-channel, not global
//   - Failed deliveries are logged but don't block the pipeline
package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/obsygnal/obsygnal-pulse/internal/config"
	"github.com/obsygnal/obsygnal-pulse/internal/domain/detection"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// Service manages alert routing, deduplication, and delivery.
type Service struct {
	alertRepo   ports.AlertRepository
	notifiers   []ports.Notifier
	log         *logger.Logger
	cfg         config.AlertingConfig

	// Deduplication: track last alert time per key
	dedupMu   sync.RWMutex
	lastAlert map[string]time.Time

	// Rate limiting: track alert count per channel per minute
	rateMu    sync.Mutex
	rateCount map[string]int
	rateReset time.Time

	// Stats
	totalFired     int64
	totalSuppressed int64
	totalDelivered int64
	totalFailed    int64
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

// HandleInefficiency processes a detected inefficiency and fires alerts if appropriate.
// This is registered as a listener on the detection engine.
func (s *Service) HandleInefficiency(ctx context.Context, ineff detection.Inefficiency) {
	// Only alert on high/critical severity
	if ineff.Severity != detection.SeverityHigh && ineff.Severity != detection.SeverityCritical {
		return
	}

	// Check deduplication
	dedupKey := fmt.Sprintf("%s:%s:%s", ineff.Type, ineff.GPUID, ineff.NodeID)
	if s.isDuplicate(dedupKey) {
		s.totalSuppressed++
		s.log.Debug().Str("key", dedupKey).Msg("alert suppressed (deduplicated)")
		return
	}

	// Create alert record
	alert := &ports.Alert{
		ID:        uuid.New().String(),
		Type:      string(ineff.Type),
		Severity:  string(ineff.Severity),
		Title:     fmt.Sprintf("[%s] %s", ineff.Severity, ineff.Description),
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

	// Dispatch to notification channels
	notification := ports.Notification{
		Title:    alert.Title,
		Message:  alert.Message,
		Severity: alert.Severity,
		Fields: map[string]string{
			"Cluster":   ineff.ClusterID,
			"Node":      ineff.NodeID,
			"GPU":       ineff.GPUID,
			"Type":      string(ineff.Type),
			"Value":     fmt.Sprintf("%.1f", ineff.CurrentValue),
			"Threshold": fmt.Sprintf("%.1f", ineff.ThresholdValue),
		},
	}

	channels := make([]string, 0)
	for _, notifier := range s.notifiers {
		if !s.isRateLimited(notifier.Type()) {
			if err := notifier.Send(ctx, notification); err != nil {
				s.totalFailed++
				s.log.Error().Err(err).Str("channel", notifier.Type()).Msg("failed to deliver alert")
			} else {
				s.totalDelivered++
				channels = append(channels, notifier.Type())
				s.log.Info().
					Str("alert_id", alert.ID).
					Str("channel", notifier.Type()).
					Str("severity", alert.Severity).
					Msg("alert delivered")
			}
		} else {
			s.totalSuppressed++
			s.log.Warn().Str("channel", notifier.Type()).Msg("alert suppressed (rate limited)")
		}
	}

	alert.Channels = channels
	s.markAlerted(dedupKey)
}

// isDuplicate checks if an alert was recently sent for this key.
func (s *Service) isDuplicate(key string) bool {
	s.dedupMu.RLock()
	defer s.dedupMu.RUnlock()

	if last, ok := s.lastAlert[key]; ok {
		return time.Since(last) < s.cfg.DeduplicationWindow
	}
	return false
}

// markAlerted records that an alert was sent for this key.
func (s *Service) markAlerted(key string) {
	s.dedupMu.Lock()
	defer s.dedupMu.Unlock()
	s.lastAlert[key] = time.Now()
}

// isRateLimited checks if a channel has exceeded its rate limit.
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

// ListRecent returns recent alerts.
func (s *Service) ListRecent(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	return s.alertRepo.ListRecent(ctx, clusterID, limit)
}

// Stats returns alerting statistics.
func (s *Service) Stats() map[string]int64 {
	return map[string]int64{
		"total_fired":      s.totalFired,
		"total_suppressed": s.totalSuppressed,
		"total_delivered":  s.totalDelivered,
		"total_failed":     s.totalFailed,
	}
}
