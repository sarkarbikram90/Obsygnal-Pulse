// Package ports — detection and recommendation port interfaces.
package ports

import (
	"context"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/detection"
	"github.com/obsygnal/obsygnal-pulse/internal/domain/recommendation"
)

// InefficiencyRepository defines persistence for detected inefficiencies.
type InefficiencyRepository interface {
	Insert(ctx context.Context, ineff *detection.Inefficiency) error
	InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error
	GetByID(ctx context.Context, id string) (*detection.Inefficiency, error)
	ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error)
	ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error)
	Resolve(ctx context.Context, id string, resolvedAt time.Time) error
	CountActive(ctx context.Context, clusterID string) (int, error)
}

// RecommendationRepository defines persistence for optimization recommendations.
type RecommendationRepository interface {
	Insert(ctx context.Context, rec *recommendation.Recommendation) error
	InsertBatch(ctx context.Context, recs []recommendation.Recommendation) error
	GetByID(ctx context.Context, id string) (*recommendation.Recommendation, error)
	ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error)
	Dismiss(ctx context.Context, id, dismissedBy, reason string) error
	Apply(ctx context.Context, id string) error
	CountActive(ctx context.Context, clusterID string) (int, error)
}

// AlertRepository defines persistence for alert history.
type AlertRepository interface {
	Insert(ctx context.Context, alert *Alert) error
	ListRecent(ctx context.Context, clusterID string, limit int) ([]Alert, error)
	Acknowledge(ctx context.Context, id, acknowledgedBy string) error
}

// Alert represents an alert record.
type Alert struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Severity       string            `json:"severity"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
	ClusterID      string            `json:"cluster_id"`
	NodeID         string            `json:"node_id,omitempty"`
	GPUID          string            `json:"gpu_id,omitempty"`
	JobID          string            `json:"job_id,omitempty"`
	FiredAt        time.Time         `json:"fired_at"`
	ResolvedAt     *time.Time        `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time        `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string            `json:"acknowledged_by,omitempty"`
	Channels       []string          `json:"channels"`
	Labels         map[string]string `json:"labels,omitempty"`
}
