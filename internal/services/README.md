# Application Services Layer (`internal/services/`)

## Overview
The `services` directory contains the core business logic of Obsygnal-Pulse. Following Domain-Driven Design (DDD) and Hexagonal Architecture principles, these services orchestrate the flow of data between incoming requests (from the API or background workers) and the outgoing infrastructure adapters (database, cache, external APIs).

## Responsibilities
- Implementing core business use cases (e.g., detecting inefficiencies, generating recommendations).
- Orchestrating multiple domain entities.
- Enforcing business rules and transaction boundaries.
- Interfacing strictly via `ports` (interfaces), never directly coupling to infrastructure.

## Architecture Role
The Service layer sits in the center of the Hexagonal Architecture, just outside the `domain` layer. It is invoked by primary adapters (e.g., HTTP Handlers in `internal/api/` or Background Workers in `cmd/worker/`) and delegates side-effects to secondary adapters (e.g., ClickHouse, Redis).

## Key Components
- `ingestion/`: Handles high-throughput GPU telemetry ingestion, validation, buffering, and batching to ClickHouse.
- `detection/`: Evaluates telemetry against defined threshold rules to identify specific inefficiencies (Idle, CPU Starvation, Memory Leaks).
- `recommendation/`: Generates actionable optimization tasks (e.g., "Downscale to T4") based on detected inefficiencies.
- `alerting/`: Routes critical incidents to external channels (Slack, Webhooks) with deduplication logic.
- `auth/`: Handles authentication and authorization validation (future expansion for RBAC).
- `analytics/`: Aggregates historical cluster data for the frontend dashboards.

## Dependencies
- **Internal**: `internal/domain` (models), `internal/ports` (interfaces), `internal/pkg/logger` (structured logging).
- **External**: None. This layer is entirely infrastructure-agnostic.

## Data Flow
```text
API Handler / Worker → Service (Business Logic) → Ports (Interfaces) → Infrastructure Adapters
```

## Operational Considerations
- **Scaling**: Services are completely stateless and can be scaled horizontally without constraint. State is pushed to Redis (for deduplication/locks) or ClickHouse (for persistence).
- **Reliability**: Services assume infrastructure can fail and rely on adapters to implement retries/circuit breaking.
- **Failure Handling**: Errors are wrapped with context using `internal/pkg/errors` before being returned to the API layer.

## Security Considerations
- Services do not handle HTTP sessions or JWT parsing directly; they receive parsed context strings (e.g., `userID`) from the API layer.
- Authorization checks (e.g., "Does this user have permission to dismiss this recommendation?") occur here.

## Observability
- **Metrics**: Time spent in service functions should be tracked via OpenTelemetry spans.
- **Logging**: Emits structured JSON logs containing `component=service_name` for easy log aggregation.

## Testing
- **Unit Tests**: 100% mocked infrastructure. We use interface mocking (e.g., `MockMetricsRepository`) to test business logic in isolation.
- **Expected Coverage**: > 85% for business critical paths.

## Future Improvements
- Integrate OpenTelemetry context propagation across all service method signatures.
- Implement an event bus (e.g., NATS/Kafka) for inter-service communication rather than direct method calls, particularly between `detection` and `recommendation`.
