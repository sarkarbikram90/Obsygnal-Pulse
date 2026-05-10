# Application Entrypoints (`cmd/`)

## Overview
The `cmd/` directory contains the main applications for the Obsygnal-Pulse project. Each subdirectory here represents a distinct deployable binary.

## Responsibilities
- Bootstrapping the application.
- Loading and validating environment configurations.
- Initializing structured logging and observability hooks (tracing/metrics).
- Establishing connections to infrastructure (Redis, ClickHouse).
- Wiring up the Hexagonal Architecture (injecting Adapters into Services).
- Managing OS signals for graceful shutdown.

## Architecture Role
This is the composition root. It is the only place in the application where concrete infrastructure adapters, generic business services, and delivery mechanisms (like HTTP routers) are permitted to meet and be wired together.

## Key Components
- `api-server/`: The public-facing REST API. Serves dashboard requests and handles high-throughput telemetry ingestion.
- `worker/`: A background daemon responsible for asynchronously evaluating detection rules against telemetry and generating optimization recommendations.
- `collector/`: (Future/WIP) An agent designed to run natively on GPU nodes to scrape local NVML metrics and push them to the `api-server`.
- `simulator/`: A developer tool that generates realistic, synthetic GPU telemetry for testing the platform without requiring actual ML hardware.

## Dependencies
- **Internal**: ALL other internal packages (`domain`, `ports`, `services`, `adapters`, `api`, `config`).
- **External**: Minimal external dependencies, mostly restricted to HTTP routing or daemon management.

## Operational Considerations
- **Deployment**: Each binary is packaged into its own distinct Docker image (see `deploy/docker/`).
- **Scaling**: 
  - `api-server` scales horizontally based on HTTP load.
  - `worker` scales horizontally based on queue length or partition logic (requires Redis locks to prevent duplicate evaluation).
- **Failure Handling**: The `main.go` file must strictly handle `SIGINT` and `SIGTERM` signals to cleanly flush buffers (like the ClickHouse ingestion batch) before exiting.

## Security Considerations
- The entrypoint is responsible for securely loading secrets from the environment and passing them explicitly to adapters. Secrets should NEVER be logged during startup.

## Testing
- **Unit Tests**: `main.go` files are typically not unit tested due to their heavy reliance on infrastructure wiring.
- **Integration Tests**: The compilation and successful bootstrapping of these binaries are validated via End-to-End (E2E) pipeline tests.

## Future Improvements
- Refactor dependency injection to use a robust container (e.g., `go.uber.org/fx` or Google `wire`) to reduce the boilerplate manual wiring currently present in the `main` functions.
