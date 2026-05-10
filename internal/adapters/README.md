# Infrastructure Adapters Layer (`internal/adapters/`)

## Overview
The `adapters` directory contains the concrete implementations of the interfaces defined in `internal/ports/`. This layer translates the application's generic domain requests into specific technology protocols (e.g., SQL queries, Redis commands, HTTP requests).

## Responsibilities
- Executing SQL queries against ClickHouse.
- Managing Redis connections, caching, and distributed locks.
- Formatting and sending payloads to external APIs (Slack, Webhooks).
- Managing connection pools, retries, and timeouts.
- Mapping database-specific structures back to pure `domain` models.

## Architecture Role
This is the outermost layer of the backend Hexagonal Architecture (Secondary Adapters). No other layer should import anything from `adapters`. Dependencies always flow *inwards* (i.e., `main.go` wires an adapter to a service, but a service never imports an adapter).

## Key Components
- `clickhouse/`: OLAP database adapter for high-volume metric persistence.
- `redis/`: In-memory data store adapter for rate limiting, caching, and deduplication.
- `slack/`: HTTP client adapter for routing alerts to Slack.
- `webhook/`: Generic HTTP adapter for outbound integrations.
- `prometheus/`: Pull-based metrics exposure.
- `kubernetes/`: Future adapter for querying native k8s resource limits.

## Dependencies
- **Internal**: `internal/ports`, `internal/domain`, `internal/config`, `internal/pkg/logger`.
- **External**: Database drivers (`github.com/ClickHouse/clickhouse-go/v2`, `github.com/redis/go-redis/v9`).

## Operational Considerations
- **Scaling**: Adapters maintain connection pools. Pool sizes must be tuned based on horizontal scaling to prevent connection exhaustion at the DB layer.
- **Reliability**: All network calls MUST use `context.Context` for timeouts. Circuit breakers should be implemented here to prevent cascading failures if a downstream service degrades.
- **Failure Handling**: Transient errors (e.g., network blips) should implement exponential backoff retries.

## Security Considerations
- Credentials (e.g., passwords, API tokens) are strictly passed in via the `config` object and must never be hardcoded.
- TLS must be enforced for ClickHouse and Redis connections in production.

## Observability
- **Metrics**: Query latency, pool exhaustion, and error rates should be tracked using Prometheus.
- **Tracing**: Redis and ClickHouse clients must propagate OpenTelemetry span context.

## Testing
- **Integration Tests**: Adapters are tested using real infrastructure spun up via Testcontainers (Docker). We do NOT use mocks to test adapters.
- **Expected Coverage**: > 80% with a focus on failure states (e.g., simulating DB disconnects).

## Future Improvements
- Implement robust Circuit Breaker patterns (e.g., using `sony/gobreaker`) on all outgoing HTTP and DB calls.
- Add OpenTelemetry hooks to the ClickHouse driver.
