// Package clickhouse implements the MetricsRepository port for the ClickHouse OLAP database.
// This secondary adapter is optimized for extreme write-throughput (batch insertions)
// and rapid time-series analytical queries. It maps raw SQL data into
// domain-native structs and strictly handles connection pooling and query timeouts.
package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// Adapter implements ports.MetricsRepository for ClickHouse.
type Adapter struct {
	conn driver.Conn
	log  *logger.Logger
}

// NewAdapter creates a new ClickHouse adapter.
func NewAdapter(cfg config.ClickHouseConfig, log *logger.Logger) (*Adapter, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer pingCancel()

	if err := conn.Ping(pingCtx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Database).
		Msg("connected to ClickHouse")

	return &Adapter{conn: conn, log: log.WithComponent("clickhouse")}, nil
}

// InsertGPUMetrics batch-inserts GPU metrics using ClickHouse native batch API.
func (a *Adapter) InsertGPUMetrics(ctx context.Context, gpuMetrics []metrics.GPUMetric) error {
	if len(gpuMetrics) == 0 {
		return nil
	}

	batch, err := a.conn.PrepareBatch(ctx, `INSERT INTO gpu_metrics (
		id, timestamp, cluster_id, node_id, gpu_id, gpu_model, gpu_index,
		utilization, memory_utilization,
		memory_used, memory_total, memory_free,
		temperature, power_usage, power_limit, fan_speed,
		pcie_tx_bytes, pcie_rx_bytes,
		running_processes, job_id, source, labels
	)`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, m := range gpuMetrics {
		source := "prometheus"
		if m.Labels != nil {
			if s, ok := m.Labels["source"]; ok {
				source = s
			}
		}
		err := batch.Append(
			m.ID, m.Timestamp, m.ClusterID, m.NodeID, m.GPUID, m.GPUModel, uint8(m.GPUIndex),
			float32(m.Utilization), float32(m.MemoryUtilization),
			m.MemoryUsed, m.MemoryTotal, m.MemoryFree,
			float32(m.Temperature), float32(m.PowerUsage), float32(m.PowerLimit), float32(m.FanSpeed),
			m.PCIeTxBytes, m.PCIeRxBytes,
			uint16(m.RunningProcesses), m.JobID, source, m.Labels,
		)
		if err != nil {
			a.log.Warn().Err(err).Str("gpu_id", m.GPUID).Msg("failed to append metric to batch")
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	a.log.Debug().Int("count", len(gpuMetrics)).Msg("inserted GPU metrics batch")
	return nil
}

// InsertNodeMetrics batch-inserts node metrics.
func (a *Adapter) InsertNodeMetrics(ctx context.Context, nodeMetrics []metrics.NodeMetric) error {
	if len(nodeMetrics) == 0 {
		return nil
	}

	batch, err := a.conn.PrepareBatch(ctx, `INSERT INTO node_metrics (
		timestamp, cluster_id, node_id,
		cpu_usage, memory_usage, disk_usage,
		network_in, network_out,
		gpu_count, pod_count, labels
	)`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, m := range nodeMetrics {
		err := batch.Append(
			m.Timestamp, m.ClusterID, m.NodeID,
			float32(m.CPUUsage), float32(m.MemoryUsage), float32(m.DiskUsage),
			m.NetworkIn, m.NetworkOut,
			uint8(m.GPUCount), uint16(m.PodCount), m.Labels,
		)
		if err != nil {
			a.log.Warn().Err(err).Str("node_id", m.NodeID).Msg("failed to append node metric")
			continue
		}
	}

	return batch.Send()
}

// QueryGPUMetrics retrieves GPU metrics matching the query.
func (a *Adapter) QueryGPUMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.GPUMetric, error) {
	sql := `SELECT
		id, timestamp, cluster_id, node_id, gpu_id, gpu_model, gpu_index,
		utilization, memory_utilization,
		memory_used, memory_total, memory_free,
		temperature, power_usage, power_limit, fan_speed,
		pcie_tx_bytes, pcie_rx_bytes,
		running_processes, job_id, labels
	FROM gpu_metrics
	WHERE 1=1`

	args := make([]interface{}, 0)

	if query.ClusterID != "" {
		sql += " AND cluster_id = ?"
		args = append(args, query.ClusterID)
	}
	if query.NodeID != "" {
		sql += " AND node_id = ?"
		args = append(args, query.NodeID)
	}
	if query.GPUID != "" {
		sql += " AND gpu_id = ?"
		args = append(args, query.GPUID)
	}
	if query.JobID != "" {
		sql += " AND job_id = ?"
		args = append(args, query.JobID)
	}
	if !query.From.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.From)
	}
	if !query.To.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.To)
	}

	sql += " ORDER BY timestamp DESC"

	if query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", query.Limit)
	} else {
		sql += " LIMIT 1000"
	}
	if query.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", query.Offset)
	}

	rows, err := a.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var result []metrics.GPUMetric
	for rows.Next() {
		var m metrics.GPUMetric
		var gpuIndex uint8
		var util, memUtil, temp, power, powerLimit, fan float32
		var procs uint16

		if err := rows.Scan(
			&m.ID, &m.Timestamp, &m.ClusterID, &m.NodeID, &m.GPUID, &m.GPUModel, &gpuIndex,
			&util, &memUtil,
			&m.MemoryUsed, &m.MemoryTotal, &m.MemoryFree,
			&temp, &power, &powerLimit, &fan,
			&m.PCIeTxBytes, &m.PCIeRxBytes,
			&procs, &m.JobID, &m.Labels,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		m.GPUIndex = int(gpuIndex)
		m.Utilization = float64(util)
		m.MemoryUtilization = float64(memUtil)
		m.Temperature = float64(temp)
		m.PowerUsage = float64(power)
		m.PowerLimit = float64(powerLimit)
		m.FanSpeed = float64(fan)
		m.RunningProcesses = int(procs)

		result = append(result, m)
	}

	return result, nil
}

// QueryNodeMetrics retrieves node metrics matching the query.
func (a *Adapter) QueryNodeMetrics(ctx context.Context, query ports.MetricsQuery) ([]metrics.NodeMetric, error) {
	sql := `SELECT timestamp, cluster_id, node_id, cpu_usage, memory_usage, disk_usage,
		network_in, network_out, gpu_count, pod_count, labels
	FROM node_metrics WHERE 1=1`

	args := make([]interface{}, 0)
	if query.ClusterID != "" {
		sql += " AND cluster_id = ?"
		args = append(args, query.ClusterID)
	}
	if query.NodeID != "" {
		sql += " AND node_id = ?"
		args = append(args, query.NodeID)
	}
	if !query.From.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.From)
	}
	if !query.To.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.To)
	}
	sql += " ORDER BY timestamp DESC LIMIT 1000"

	rows, err := a.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var result []metrics.NodeMetric
	for rows.Next() {
		var m metrics.NodeMetric
		var cpu, mem, disk float32
		var gpuCount uint8
		var podCount uint16
		if err := rows.Scan(
			&m.Timestamp, &m.ClusterID, &m.NodeID, &cpu, &mem, &disk,
			&m.NetworkIn, &m.NetworkOut, &gpuCount, &podCount, &m.Labels,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		m.CPUUsage = float64(cpu)
		m.MemoryUsage = float64(mem)
		m.DiskUsage = float64(disk)
		m.GPUCount = int(gpuCount)
		m.PodCount = int(podCount)
		result = append(result, m)
	}
	return result, nil
}

// GetGPUUtilizationAvg returns the average GPU utilization over a time range.
func (a *Adapter) GetGPUUtilizationAvg(ctx context.Context, gpuID string, from, to time.Time) (float64, error) {
	var avg float64
	err := a.conn.QueryRow(ctx,
		`SELECT avg(utilization) FROM gpu_metrics WHERE gpu_id = ? AND timestamp >= ? AND timestamp <= ?`,
		gpuID, from, to,
	).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to get GPU utilization avg: %w", err)
	}
	return avg, nil
}

// GetClusterOverview returns aggregated cluster statistics.
func (a *Adapter) GetClusterOverview(ctx context.Context, clusterID string) (*ports.ClusterOverview, error) {
	overview := &ports.ClusterOverview{
		ClusterID: clusterID,
		Timestamp: time.Now(),
	}

	// Get node count
	err := a.conn.QueryRow(ctx,
		`SELECT uniqExact(node_id) FROM node_metrics
		 WHERE cluster_id = ? AND timestamp >= now() - INTERVAL 5 MINUTE`,
		clusterID,
	).Scan(&overview.TotalNodes)
	if err != nil {
		a.log.Warn().Err(err).Msg("failed to get node count")
	}
	overview.HealthyNodes = overview.TotalNodes // simplified for MVP

	// Get GPU stats
	err = a.conn.QueryRow(ctx,
		`SELECT
			uniqExact(gpu_id),
			countIf(utilization > 5),
			countIf(utilization <= 5),
			avg(utilization),
			avg(memory_utilization)
		FROM gpu_metrics
		WHERE cluster_id = ? AND timestamp >= now() - INTERVAL 5 MINUTE`,
		clusterID,
	).Scan(
		&overview.TotalGPUs,
		&overview.ActiveGPUs,
		&overview.IdleGPUs,
		&overview.AvgUtilization,
		&overview.AvgMemoryUsage,
	)
	if err != nil {
		a.log.Warn().Err(err).Msg("failed to get GPU stats")
	}

	// Calculate inefficiency score (simplified: 100 - avg utilization)
	overview.InefficiencyScore = 100 - overview.AvgUtilization

	// Estimate waste (simplified: idle GPUs * $2.50/hr A100 rate)
	overview.EstimatedWaste = float64(overview.IdleGPUs) * 2.50

	return overview, nil
}

// Ping checks the ClickHouse connection.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.conn.Ping(ctx)
}

// Conn returns the raw ClickHouse connection.
func (a *Adapter) Conn() driver.Conn {
	return a.conn
}

// Close closes the ClickHouse connection.
func (a *Adapter) Close() error {
	return a.conn.Close()
}
