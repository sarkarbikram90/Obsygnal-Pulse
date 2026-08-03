package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// AlertAdapter implements ports.AlertRepository for ClickHouse.
type AlertAdapter struct {
	conn driver.Conn
	log  *logger.Logger
}

// NewAlertAdapter creates a new AlertAdapter.
func NewAlertAdapter(conn driver.Conn, log *logger.Logger) *AlertAdapter {
	return &AlertAdapter{
		conn: conn,
		log:  log.WithComponent("clickhouse-alert"),
	}
}

func (a *AlertAdapter) Insert(ctx context.Context, alert *ports.Alert) error {
	batch, err := a.conn.PrepareBatch(ctx, `
		INSERT INTO alerts (
			id, type, severity, title, message, cluster_id, node_id, gpu_id, job_id,
			fired_at, channels, labels
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare alert batch: %w", err)
	}

	err = batch.Append(
		alert.ID, alert.Type, alert.Severity, alert.Title, alert.Message,
		alert.ClusterID, alert.NodeID, alert.GPUID, alert.JobID,
		alert.FiredAt, alert.Channels, alert.Labels,
	)
	if err != nil {
		return fmt.Errorf("failed to append alert: %w", err)
	}

	return batch.Send()
}

func (a *AlertAdapter) ListRecent(ctx context.Context, clusterID string, limit int) ([]ports.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := a.conn.Query(ctx, `
		SELECT id, type, severity, title, message, cluster_id, node_id, gpu_id, job_id,
			fired_at, resolved_at, acknowledged_at, acknowledged_by, channels, labels
		FROM alerts
		WHERE cluster_id = ?
		ORDER BY fired_at DESC
		LIMIT ?
	`, clusterID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent alerts: %w", err)
	}
	defer rows.Close()

	var alerts []ports.Alert
	for rows.Next() {
		var alert ports.Alert
		err := rows.Scan(
			&alert.ID, &alert.Type, &alert.Severity, &alert.Title, &alert.Message,
			&alert.ClusterID, &alert.NodeID, &alert.GPUID, &alert.JobID,
			&alert.FiredAt, &alert.ResolvedAt, &alert.AcknowledgedAt, &alert.AcknowledgedBy,
			&alert.Channels, &alert.Labels,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (a *AlertAdapter) Acknowledge(ctx context.Context, id, acknowledgedBy string) error {
	now := time.Now()
	return a.conn.Exec(ctx, `
		ALTER TABLE alerts UPDATE acknowledged_at = ?, acknowledged_by = ? WHERE id = ?
	`, now, acknowledgedBy, id)
}
