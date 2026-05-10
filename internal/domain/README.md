# Domain Layer (`internal/domain/`)

## Overview
The `domain` directory is the absolute core of Obsygnal-Pulse. It contains the fundamental business models, types, and logic that define what the system *is*. This layer contains pure Go code and is strictly isolated from any external concerns (no HTTP, no SQL, no JSON tags if possible).

## Responsibilities
- Defining core entities (e.g., `GPUMetric`, `Inefficiency`, `Recommendation`, `Alert`).
- Encapsulating pure business rules (e.g., determining if a specific GPU state constitutes "idle").
- Providing domain-specific methods on entities.

## Architecture Role
This is the innermost layer of the Hexagonal Architecture. 
**Crucial Rule:** The `domain` package MUST NOT import any other package within `obsygnal-pulse` (except possibly `internal/pkg/errors` or standard library utilities). It has zero dependencies on infrastructure, databases, or frameworks.

## Key Components
- `metrics/`: Models representing raw GPU telemetry (utilization, memory, temperature, power).
- `detection/`: Models representing identified inefficiencies (e.g., `CPUStarvation`, `MemoryLeak`).
- `recommendation/`: Models representing actionable optimizations proposed to the user.
- `alerting/`: Models for notification routing.
- `cluster/`: Models representing the physical topology of the ML infrastructure (Nodes, Racks, Clusters).

## Dependencies
- **Internal**: None.
- **External**: Standard Library only (`time`, `strings`, etc.).

## Data Flow
Domain objects are instantiated by incoming Primary Adapters (like the API layer) or fetched by Secondary Adapters (like the DB layer), and then passed into the Service layer where business workflows are executed.

## Operational Considerations
- Domain logic is executed entirely in memory. It must be highly optimized to prevent excessive heap allocations, especially when parsing large batches of telemetry.
- **Reliability**: Pure functions in the domain layer cannot fail due to network partitions, ensuring business logic rules remain deterministic.

## Security Considerations
- Data classification rules (e.g., stripping sensitive job IDs or PII) should be codified as domain rules on the entities themselves before they reach the infrastructure layer for persistence.

## Observability
- Since this layer contains pure business logic, tracing is typically not embedded here to avoid polluting the domain models with OpenTelemetry packages.

## Testing
- **Unit Tests**: Domain tests should be extremely fast, deterministic, and require zero mocking.
- **Expected Coverage**: 100%. This is the most critical logic in the application.

## Future Improvements
- Transition to a more robust DDD pattern by introducing strictly enforced Aggregates and Value Objects.
