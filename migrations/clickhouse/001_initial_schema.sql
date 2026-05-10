-- Obsygnal-Pulse ClickHouse Schema
-- Migration: 001_initial_schema
--
-- This schema establishes the foundational OLAP data structures for the platform.
-- It defines the core MergeTree tables required to ingest massive streams of raw GPU
-- telemetry (gpu_metrics) and Node statistics. It leverages LowCardinality for
-- high-cardinality metadata to aggressively reduce RAM consumption and I/O.
-- It also defines the Materialized Views (gpu_metrics_hourly_mv) that pre-aggregate
-- raw telemetry into time-bucketed rollups, drastically reducing query latency for
-- the Grafana dashboards and preventing full-table scans during recommendation generation.

-- ============================================================
-- RAW GPU METRICS
-- ============================================================
CREATE TABLE IF NOT EXISTS gpu_metrics (
    -- Identity
    id              String,
    timestamp       DateTime64(3, 'UTC'),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    gpu_id          String,
    gpu_model       LowCardinality(String),
    gpu_index       UInt8,

    -- Utilization (0-100)
    utilization           Float32,
    memory_utilization    Float32,

    -- Memory (bytes)
    memory_used     UInt64,
    memory_total    UInt64,
    memory_free     UInt64,

    -- Thermal and power
    temperature     Float32,
    power_usage     Float32,
    power_limit     Float32,
    fan_speed       Float32,

    -- PCIe
    pcie_tx_bytes   UInt64,
    pcie_rx_bytes   UInt64,

    -- Process info
    running_processes UInt16,
    job_id          String DEFAULT '',

    -- Metadata
    source          LowCardinality(String) DEFAULT 'prometheus',
    labels          Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (cluster_id, node_id, gpu_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- NODE METRICS
-- ============================================================
CREATE TABLE IF NOT EXISTS node_metrics (
    timestamp       DateTime64(3, 'UTC'),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    cpu_usage       Float32,
    memory_usage    Float32,
    disk_usage      Float32,
    network_in      UInt64,
    network_out     UInt64,
    gpu_count       UInt8,
    pod_count       UInt16,
    labels          Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (cluster_id, node_id, timestamp)
TTL toDateTime(timestamp) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- TRAINING JOBS
-- ============================================================
CREATE TABLE IF NOT EXISTS training_jobs (
    id              String,
    name            String,
    namespace       LowCardinality(String),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    status          LowCardinality(String),
    gpu_ids         Array(String),
    gpu_count       UInt8,
    framework       LowCardinality(String),
    started_at      DateTime64(3, 'UTC'),
    completed_at    Nullable(DateTime64(3, 'UTC')),
    runtime_seconds Float64,
    queue_wait_seconds Float64,
    retry_count     UInt16,
    exit_code       Int32,
    error_msg       String DEFAULT '',
    requested_gpus  UInt8,
    requested_memory UInt64,
    requested_cpus  Float32,
    avg_gpu_utilization Float32,
    max_gpu_utilization Float32,
    avg_memory_usage    Float32,
    max_memory_usage    Float32,
    labels          Map(String, String),
    updated_at      DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(started_at)
ORDER BY (cluster_id, id)
TTL toDateTime(started_at) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- INEFFICIENCIES
-- ============================================================
CREATE TABLE IF NOT EXISTS inefficiencies (
    id              String,
    type            LowCardinality(String),
    severity        LowCardinality(String),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    gpu_id          String DEFAULT '',
    job_id          String DEFAULT '',
    description     String,
    affected_metrics Array(String),
    current_value   Float64,
    threshold_value Float64,
    detected_at     DateTime64(3, 'UTC'),
    resolved_at     Nullable(DateTime64(3, 'UTC')),
    is_active       UInt8,
    labels          Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(detected_at)
ORDER BY (cluster_id, type, detected_at)
TTL toDateTime(detected_at) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- RECOMMENDATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS recommendations (
    id              String,
    category        LowCardinality(String),
    status          LowCardinality(String),
    title           String,
    description     String,
    explanation     String,
    confidence_score Float32,
    cost_per_hour_usd   Float64,
    cost_per_month_usd  Float64,
    gpu_hours_per_month  Float64,
    affected_metrics     Array(String),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String) DEFAULT '',
    gpu_id          String DEFAULT '',
    job_id          String DEFAULT '',
    inefficiency_id String,
    created_at      DateTime64(3, 'UTC'),
    expires_at      DateTime64(3, 'UTC'),
    dismissed_at    Nullable(DateTime64(3, 'UTC')),
    dismissed_by    String DEFAULT '',
    dismiss_reason  String DEFAULT '',
    labels          Map(String, String),
    updated_at      DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_at)
ORDER BY (cluster_id, id)
TTL toDateTime(created_at) + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- ALERTS
-- ============================================================
CREATE TABLE IF NOT EXISTS alerts (
    id              String,
    type            LowCardinality(String),
    severity        LowCardinality(String),
    title           String,
    message         String,
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String) DEFAULT '',
    gpu_id          String DEFAULT '',
    job_id          String DEFAULT '',
    fired_at        DateTime64(3, 'UTC'),
    resolved_at     Nullable(DateTime64(3, 'UTC')),
    acknowledged_at Nullable(DateTime64(3, 'UTC')),
    acknowledged_by String DEFAULT '',
    channels        Array(String),
    labels          Map(String, String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(fired_at)
ORDER BY (cluster_id, type, fired_at)
TTL toDateTime(fired_at) + INTERVAL 180 DAY
SETTINGS index_granularity = 8192;


-- ============================================================
-- MATERIALIZED VIEWS FOR AGGREGATION
-- ============================================================

-- Hourly GPU utilization aggregates
CREATE TABLE IF NOT EXISTS gpu_metrics_hourly (
    timestamp_hour  DateTime('UTC'),
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    gpu_id          String,
    gpu_model       LowCardinality(String),
    avg_utilization     Float64,
    max_utilization     Float64,
    min_utilization     Float64,
    avg_memory_utilization Float64,
    avg_temperature     Float64,
    max_temperature     Float64,
    avg_power_usage     Float64,
    total_pcie_tx       UInt64,
    total_pcie_rx       UInt64,
    sample_count        UInt64
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(timestamp_hour)
ORDER BY (cluster_id, node_id, gpu_id, timestamp_hour)
TTL timestamp_hour + INTERVAL 365 DAY;

CREATE MATERIALIZED VIEW IF NOT EXISTS gpu_metrics_hourly_mv
TO gpu_metrics_hourly AS
SELECT
    toStartOfHour(timestamp) AS timestamp_hour,
    cluster_id,
    node_id,
    gpu_id,
    gpu_model,
    avg(utilization)        AS avg_utilization,
    max(utilization)        AS max_utilization,
    min(utilization)        AS min_utilization,
    avg(memory_utilization) AS avg_memory_utilization,
    avg(temperature)        AS avg_temperature,
    max(temperature)        AS max_temperature,
    avg(power_usage)        AS avg_power_usage,
    sum(pcie_tx_bytes)      AS total_pcie_tx,
    sum(pcie_rx_bytes)      AS total_pcie_rx,
    count()                 AS sample_count
FROM gpu_metrics
GROUP BY
    timestamp_hour,
    cluster_id,
    node_id,
    gpu_id,
    gpu_model;


-- Daily GPU utilization aggregates
CREATE TABLE IF NOT EXISTS gpu_metrics_daily (
    timestamp_day   Date,
    cluster_id      LowCardinality(String),
    node_id         LowCardinality(String),
    gpu_id          String,
    gpu_model       LowCardinality(String),
    avg_utilization     Float64,
    max_utilization     Float64,
    min_utilization     Float64,
    avg_memory_utilization Float64,
    avg_temperature     Float64,
    max_temperature     Float64,
    avg_power_usage     Float64,
    sample_count        UInt64
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(timestamp_day)
ORDER BY (cluster_id, node_id, gpu_id, timestamp_day)
TTL timestamp_day + INTERVAL 1095 DAY;

CREATE MATERIALIZED VIEW IF NOT EXISTS gpu_metrics_daily_mv
TO gpu_metrics_daily AS
SELECT
    toDate(timestamp) AS timestamp_day,
    cluster_id,
    node_id,
    gpu_id,
    gpu_model,
    avg(utilization)        AS avg_utilization,
    max(utilization)        AS max_utilization,
    min(utilization)        AS min_utilization,
    avg(memory_utilization) AS avg_memory_utilization,
    avg(temperature)        AS avg_temperature,
    max(temperature)        AS max_temperature,
    avg(power_usage)        AS avg_power_usage,
    count()                 AS sample_count
FROM gpu_metrics
GROUP BY
    timestamp_day,
    cluster_id,
    node_id,
    gpu_id,
    gpu_model;
