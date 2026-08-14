# Application Services Layer (`internal/services/`)

## Overview
The `services` directory contains the business orchestration logic of Obsygnal-Pulse. Services coordinate the flow of data between primary adapters (API handlers, background workers) and secondary adapters (ClickHouse, Redis, notification channels) strictly via port interfaces defined in `internal/ports/`.

---

## Implemented Services

| Service | Location | Responsibilities |
| :--- | :--- | :--- |
| **Ingestion** | `ingestion/` | High-throughput telemetry ingestion, buffer management, time-series deduplication, and periodic batch flushing to ClickHouse. |
| **Detection** | `detection/` | Periodic evaluation of recent telemetry against the 9-rule Inefficiency Detection Engine. |
| **Recommendation** | `recommendation/` | Translating detected inefficiencies into actionable recommendations with multi-cloud ROI monetary savings calculations. |
| **Auth & SaaS** | `auth/` | Tenant registration, user login, bcrypt password hashing, JWT token issuance, machine API key management, and SOC2 audit event logging. |
| **Alerting** | `alerting/` | Incident dispatching to Slack and generic webhooks with deduplication windows and rate limiting. |
| **Analytics** | `analytics/` | Aggregating cluster-wide utilization, node health, and total wasted spend for dashboards. |

---

## Architectural Rules

1. **Port Decoupling**: Services depend exclusively on interface definitions in `internal/ports/`. They never instantiate concrete adapters directly.
2. **Stateless Scalability**: All service state is stored either in memory buffers with explicit mutex locking or persisted to database adapters.
3. **Structured Logging**: Services emit structured logs via `internal/pkg/logger` including `component` and `service` tags for unified distributed tracing.

---

## Testing Services

```bash
go test -v ./internal/services/...
```
Service unit tests use in-memory test doubles and assert business rules without external infrastructure.
