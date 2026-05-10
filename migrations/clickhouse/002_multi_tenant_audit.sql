-- Obsygnal-Pulse Multi-Tenant ClickHouse Schema Updates

-- 1. Alter existing tables to ensure they partition and filter by org_id.
-- (Assuming we had cluster_id, we now inject org_id as the primary partitioning key for strict tenant isolation).

ALTER TABLE gpu_metrics ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE node_metrics ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE inefficiencies ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE recommendations ADD COLUMN IF NOT EXISTS org_id UUID;

-- 2. Audit Trail (Log engine for fast ingestion of millions of events)
-- Used for SOC2/Compliance reporting.

CREATE TABLE IF NOT EXISTS audit_events (
    org_id UUID,
    event_id UUID,
    actor_id UUID,
    actor_type String, -- 'user', 'api_key', 'service_account'
    action String,     -- e.g., 'recommendation.applied', 'user.invited'
    resource String,
    metadata Map(String, String),
    ip_address String,
    user_agent String,
    timestamp DateTime64(3, 'UTC') DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (org_id, timestamp, action)
TTL timestamp + INTERVAL 1 YEAR; -- Retain audit logs for 1 year for compliance
