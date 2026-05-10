package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// RecommendationAdapter implements the ports.RecommendationRepository interface for ClickHouse.
type RecommendationAdapter struct {
	conn driver.Conn
	log  *logger.Logger
}

// NewRecommendationAdapter creates a new RecommendationAdapter.
func NewRecommendationAdapter(conn driver.Conn, log *logger.Logger) *RecommendationAdapter {
	return &RecommendationAdapter{
		conn: conn,
		log:  log.WithComponent("clickhouse-recommendation"),
	}
}

func (a *RecommendationAdapter) Insert(ctx context.Context, rec *recommendation.Recommendation) error {
	return a.InsertBatch(ctx, []recommendation.Recommendation{*rec})
}

func (a *RecommendationAdapter) InsertBatch(ctx context.Context, recs []recommendation.Recommendation) error {
	if len(recs) == 0 {
		return nil
	}

	batch, err := a.conn.PrepareBatch(ctx, `
		INSERT INTO recommendations (
			id, category, status, title, description, explanation,
			confidence_score, cost_per_hour_usd, cost_per_month_usd, gpu_hours_per_month,
			affected_metrics, cluster_id, node_id, gpu_id, job_id, inefficiency_id,
			created_at, expires_at, labels, updated_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, r := range recs {
		err := batch.Append(
			r.ID, string(r.Category), string(r.Status), r.Title, r.Description, r.Explanation,
			r.ConfidenceScore, r.EstimatedSavings.CostPerHourUSD, r.EstimatedSavings.CostPerMonthUSD, r.EstimatedSavings.GPUHoursPerMonth,
			r.AffectedMetrics, r.ClusterID, r.NodeID, r.GPUID, r.JobID, r.InefficiencyID,
			r.CreatedAt, r.ExpiresAt, r.Labels, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to append recommendation to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func (a *RecommendationAdapter) GetByID(ctx context.Context, id string) (*recommendation.Recommendation, error) {
	// Not fully implemented for MVP
	return nil, nil
}

func (a *RecommendationAdapter) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	rows, err := a.conn.Query(ctx, `
		SELECT id, category, status, title, description, explanation,
			confidence_score, cost_per_hour_usd, cost_per_month_usd, gpu_hours_per_month,
			affected_metrics, cluster_id, node_id, gpu_id, job_id, inefficiency_id,
			created_at, expires_at, labels
		FROM recommendations
		WHERE cluster_id = ? AND status = 'active' AND expires_at > now()
		ORDER BY confidence_score DESC, cost_per_month_usd DESC
		LIMIT ? OFFSET ?
	`, clusterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list active recommendations: %w", err)
	}
	defer rows.Close()

	var recs []recommendation.Recommendation
	for rows.Next() {
		var rec recommendation.Recommendation
		err := rows.Scan(
			&rec.ID, &rec.Category, &rec.Status, &rec.Title, &rec.Description, &rec.Explanation,
			&rec.ConfidenceScore, &rec.EstimatedSavings.CostPerHourUSD, &rec.EstimatedSavings.CostPerMonthUSD, &rec.EstimatedSavings.GPUHoursPerMonth,
			&rec.AffectedMetrics, &rec.ClusterID, &rec.NodeID, &rec.GPUID, &rec.JobID, &rec.InefficiencyID,
			&rec.CreatedAt, &rec.ExpiresAt, &rec.Labels,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func (a *RecommendationAdapter) Dismiss(ctx context.Context, id, dismissedBy, reason string) error {
	err := a.conn.Exec(ctx, `
		ALTER TABLE recommendations UPDATE status = 'dismissed', dismissed_by = ?, dismiss_reason = ?, dismissed_at = now(), updated_at = now() WHERE id = ?
	`, dismissedBy, reason, id)
	return err
}

func (a *RecommendationAdapter) Apply(ctx context.Context, id string) error {
	err := a.conn.Exec(ctx, `
		ALTER TABLE recommendations UPDATE status = 'applied', updated_at = now() WHERE id = ?
	`, id)
	return err
}

func (a *RecommendationAdapter) CountActive(ctx context.Context, clusterID string) (int, error) {
	var count uint64
	err := a.conn.QueryRow(ctx, `
		SELECT count() FROM recommendations WHERE cluster_id = ? AND status = 'active' AND expires_at > now()
	`, clusterID).Scan(&count)
	return int(count), err
}
