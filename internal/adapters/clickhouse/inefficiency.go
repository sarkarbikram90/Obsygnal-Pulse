package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// InefficiencyAdapter implements the ports.InefficiencyRepository interface for ClickHouse.
type InefficiencyAdapter struct {
	conn driver.Conn
	log  *logger.Logger
}

// NewInefficiencyAdapter creates a new InefficiencyAdapter.
func NewInefficiencyAdapter(conn driver.Conn, log *logger.Logger) *InefficiencyAdapter {
	return &InefficiencyAdapter{
		conn: conn,
		log:  log.WithComponent("clickhouse-inefficiency"),
	}
}

func (a *InefficiencyAdapter) Insert(ctx context.Context, ineff *detection.Inefficiency) error {
	return a.InsertBatch(ctx, []detection.Inefficiency{*ineff})
}

func (a *InefficiencyAdapter) InsertBatch(ctx context.Context, ineffs []detection.Inefficiency) error {
	if len(ineffs) == 0 {
		return nil
	}

	batch, err := a.conn.PrepareBatch(ctx, `
		INSERT INTO inefficiencies (
			id, type, severity, cluster_id, node_id, gpu_id, job_id,
			description, affected_metrics, current_value, threshold_value,
			detected_at, is_active, labels
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, i := range ineffs {
		isActive := uint8(0)
		if i.IsActive {
			isActive = 1
		}

		err := batch.Append(
			i.ID, string(i.Type), string(i.Severity), i.ClusterID, i.NodeID, i.GPUID, i.JobID,
			i.Description, i.AffectedMetrics, i.CurrentValue, i.ThresholdValue,
			i.DetectedAt, isActive, i.Labels,
		)
		if err != nil {
			return fmt.Errorf("failed to append inefficiency to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func (a *InefficiencyAdapter) GetByID(ctx context.Context, id string) (*detection.Inefficiency, error) {
	var ineff detection.Inefficiency
	var isActive uint8

	err := a.conn.QueryRow(ctx, `
		SELECT id, type, severity, cluster_id, node_id, gpu_id, job_id,
			description, affected_metrics, current_value, threshold_value,
			detected_at, resolved_at, is_active, labels
		FROM inefficiencies
		WHERE id = ?
		ORDER BY detected_at DESC
		LIMIT 1
	`, id).Scan(
		&ineff.ID, &ineff.Type, &ineff.Severity, &ineff.ClusterID, &ineff.NodeID, &ineff.GPUID, &ineff.JobID,
		&ineff.Description, &ineff.AffectedMetrics, &ineff.CurrentValue, &ineff.ThresholdValue,
		&ineff.DetectedAt, &ineff.ResolvedAt, &isActive, &ineff.Labels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query inefficiency: %w", err)
	}

	ineff.IsActive = isActive == 1
	return &ineff, nil
}

func (a *InefficiencyAdapter) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]detection.Inefficiency, error) {
	rows, err := a.conn.Query(ctx, `
		SELECT id, type, severity, cluster_id, node_id, gpu_id, job_id,
			description, affected_metrics, current_value, threshold_value,
			detected_at, resolved_at, is_active, labels
		FROM inefficiencies
		WHERE cluster_id = ? AND is_active = 1
		ORDER BY detected_at DESC
		LIMIT ? OFFSET ?
	`, clusterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list active inefficiencies: %w", err)
	}
	defer rows.Close()

	var ineffs []detection.Inefficiency
	for rows.Next() {
		var ineff detection.Inefficiency
		var isActive uint8
		err := rows.Scan(
			&ineff.ID, &ineff.Type, &ineff.Severity, &ineff.ClusterID, &ineff.NodeID, &ineff.GPUID, &ineff.JobID,
			&ineff.Description, &ineff.AffectedMetrics, &ineff.CurrentValue, &ineff.ThresholdValue,
			&ineff.DetectedAt, &ineff.ResolvedAt, &isActive, &ineff.Labels,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inefficiency: %w", err)
		}
		ineff.IsActive = isActive == 1
		ineffs = append(ineffs, ineff)
	}
	return ineffs, nil
}

func (a *InefficiencyAdapter) ListByType(ctx context.Context, clusterID string, itype detection.InefficiencyType, from, to time.Time) ([]detection.Inefficiency, error) {
	// Not fully implemented for MVP
	return nil, nil
}

func (a *InefficiencyAdapter) Resolve(ctx context.Context, id string, resolvedAt time.Time) error {
	// In ClickHouse, updating is done via ALTER TABLE UPDATE or ReplacingMergeTree.
	// For MVP, we insert a new row with is_active = 0 (Requires ReplacingMergeTree or argmax in views).
	// Since inefficiencies table is MergeTree, we use ALTER TABLE UPDATE (asynchronous).
	err := a.conn.Exec(ctx, `
		ALTER TABLE inefficiencies UPDATE is_active = 0, resolved_at = ? WHERE id = ?
	`, resolvedAt, id)
	return err
}

func (a *InefficiencyAdapter) CountActive(ctx context.Context, clusterID string) (int, error) {
	var count uint64
	err := a.conn.QueryRow(ctx, `
		SELECT count() FROM inefficiencies WHERE cluster_id = ? AND is_active = 1
	`, clusterID).Scan(&count)
	return int(count), err
}
