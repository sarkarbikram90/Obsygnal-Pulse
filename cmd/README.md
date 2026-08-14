# Application Entrypoints (`cmd/`)

## Overview
The `cmd/` directory houses the `main.go` entry points for the four deployable microservices in Obsygnal-Pulse. Each subpackage represents an independently compilable and runnable binary.

---

## Deployable Binaries

```
cmd/
├── api-server/    # Public REST API control plane & high-throughput telemetry ingestion
├── collector/     # Edge node daemon for scraping local DCGM / NVML hardware metrics
├── worker/        # Asynchronous background daemon evaluating detection rules & ROI
└── simulator/     # High-fidelity synthetic GPU telemetry generator for testing & demos
```

---

## Service Responsibilities & Composition

| Binary | Role | Description |
| :--- | :--- | :--- |
| **`api-server`** | API Gateway & Control Plane | Serves the React web UI, exposes REST routes, validates JWT/API key authentication, enforces Redis rate limiting, and batches telemetry into ClickHouse. |
| **`collector`** | Edge Telemetry Daemon | Runs as a Kubernetes DaemonSet on GPU nodes, scraping NVIDIA DCGM Prometheus exporters (`:9400/metrics`) and streaming batches to the `api-server`. |
| **`worker`** | Background Intelligence Daemon | Periodically evaluates recent telemetry against the 9-rule Inefficiency Detection Engine, triggers the ROI Recommendation Engine, and routes alerts. |
| **`simulator`** | Synthetic Workload Generator | Generates realistic GPU metrics (H100, A100 clusters) simulating idle GPUs, CUDA fragmentation, and thermal spikes for development and VC demos. |

---

## Architectural Principles

### 1. Composition Root
The `cmd/` packages serve as the **composition root** in Hexagonal Architecture. This is the only location where concrete infrastructure adapters (ClickHouse, Redis, Prometheus, Slack) and delivery mechanisms (Gin HTTP engine) are instantiated and injected into application services.

### 2. Zero-Dependency Resilience & Graceful Fallback
All binaries implement graceful fallback patterns:
- If **ClickHouse is unreachable** → Services seamlessly fall back to thread-safe, stateful in-memory repositories.
- If **Redis is offline** → Rate limiting fails open and uses local memory caches.
- This ensures developers can run `go run ./cmd/api-server/main.go` out of the box with zero external infrastructure required.

### 3. Graceful Shutdown
Every binary listens for `os.Interrupt`, `SIGINT`, and `SIGTERM` signals. Upon receiving a shutdown signal, they:
- Stop accepting new incoming traffic.
- Flush in-memory telemetry buffers to persistence.
- Cleanly close database connection pools and background ticker loops within a configurable timeout window (5–10s).

---

## Building Binaries

```bash
# Build all 4 binaries into bin/
make build

# Build individual binaries
go build -o bin/api-server ./cmd/api-server/
go build -o bin/collector  ./cmd/collector/
go build -o bin/worker     ./cmd/worker/
go build -o bin/simulator  ./cmd/simulator/
```
