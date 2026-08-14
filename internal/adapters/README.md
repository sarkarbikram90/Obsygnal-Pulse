# Infrastructure Adapters Layer (`internal/adapters/`)

## Overview
The `adapters` directory contains concrete implementations of the port interfaces defined in `internal/ports/`. This layer translates generic domain requests into specific technology protocols (SQL queries, Redis key-value operations, NATS publish/subscribe, HTTP webhook dispatches).

---

## Implemented Adapters

| Directory | Target Infrastructure | Purpose |
| :--- | :--- | :--- |
| `clickhouse/` | ClickHouse OLAP | High-throughput time-series metrics persistence, analytical aggregation queries, and audit logging. |
| `redis/` | Redis In-Memory Store | Sliding-window API rate limiting, distributed locking, and cache management. |
| `nats/` | NATS JetStream | Real-time event streaming and dead-letter queue (DLQ) ingestion pipeline. |
| `memory/` | In-Memory (Go) | Stateful, thread-safe fallback repositories for local development, CI testing, and zero-dependency boots. |
| `slack/` | Slack Webhooks | Rich alert payload formatting and incident notifications. |
| `webhook/` | Generic HTTP Webhooks | JSON alert notifications to external incident management systems. |
| `prometheus/` | Prometheus PromQL | Pull-based DCGM metric querying and operational telemetry scraping. |
| `kubernetes/` | Kubernetes Client | Pod inspection and node resource limits querying. |

---

## Architectural Principles

1. **Dependency Inversion**: Dependencies flow inwards. Adapters import `ports` and `domain`, but services never import `adapters`.
2. **Context Cancellation & Timeouts**: All network operations strictly accept and respect `context.Context` to prevent goroutine leaks.
3. **Resilience**: Database connection pools are actively managed with connection lifetimes, max open connections, and exponential backoff retries.
4. **Thread Safety**: All in-memory adapters protect internal state with `sync.RWMutex`.
