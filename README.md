# 🔥 Obsygnal-Pulse

**AI Infrastructure Efficiency Platform** — Detect wasted GPU compute, optimize training workloads, save money.

[![CI](https://github.com/obsygnal/obsygnal-pulse/actions/workflows/ci.yml/badge.svg)](https://github.com/obsygnal/obsygnal-pulse/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/obsygnal/obsygnal-pulse)](https://goreportcard.com/report/github.com/obsygnal/obsygnal-pulse)

---

## What is Obsygnal-Pulse?

Obsygnal-Pulse is an observability and intelligence platform that monitors GPU clusters running AI/ML workloads, detects inefficiencies (idle GPUs, underutilized memory, failed training loops), and provides actionable optimization recommendations with confidence scores and estimated cost savings.

### Key Features

- **GPU Cluster Dashboard** — Real-time cluster overview, node health, GPU utilization heatmaps, cost estimation
- **Inefficiency Detection** — Automatically detect underutilized GPUs, oversized allocations, stalled training, high retry patterns
- **Optimization Recommendations** — Actionable suggestions with confidence scores and estimated savings
- **Alerting** — Slack, webhook, threshold, and anomaly-based alerts
- **Historical Analytics** — Job history, efficiency trends, cost trends, failure analysis

---

## Architecture

- **Backend**: Go (Gin), hexagonal architecture, modular monolith
- **Frontend**: React, TypeScript, TailwindCSS, shadcn/ui, Recharts
- **Storage**: ClickHouse (analytics), Redis (cache/queues), Prometheus (real-time metrics)
- **Infrastructure**: Docker, Kubernetes, Helm, Terraform (GKE)
- **Observability**: Grafana, Loki, OpenTelemetry

---

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make

### 1. Clone & Setup

```bash
git clone https://github.com/obsygnal/obsygnal-pulse.git
cd obsygnal-pulse
cp .env.example .env
make setup
```

### 2. Start Infrastructure

```bash
make docker-up
```

This starts:
- **ClickHouse** — `localhost:8123` (HTTP) / `localhost:9000` (native)
- **Redis** — `localhost:6379`
- **Prometheus** — `http://localhost:9090`
- **Grafana** — `http://localhost:3000` (admin / obsygnal)

### 3. Run Migrations

```bash
make migrate-docker
```

### 4. Run the API Server

```bash
make dev
```

API available at `http://localhost:8080`

### 5. Run the GPU Simulator

```bash
make dev-simulator
```

Generates fake GPU telemetry for development without real GPU hardware.

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/api/v1/ready` | Readiness probe |
| `GET` | `/api/v1/cluster/overview` | Cluster dashboard data |
| `GET` | `/api/v1/gpus/metrics` | GPU metric time series |
| `POST` | `/api/v1/metrics/ingest` | Ingest metric batches |
| `GET` | `/api/v1/ingestion/stats` | Ingestion pipeline stats |

---

## Development

```bash
make help          # Show all available commands
make build         # Build all binaries
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint          # Run linters
make fmt           # Format code
```

---

## Project Structure

```
├── cmd/                    # Application entry points
│   ├── api-server/         # REST API server
│   ├── collector/          # Telemetry collector
│   ├── worker/             # Background processing
│   └── simulator/          # Fake GPU telemetry generator
├── internal/               # Private application code
│   ├── domain/             # Domain models (no external deps)
│   ├── ports/              # Interface definitions (hexagonal)
│   ├── adapters/           # Infrastructure implementations
│   ├── services/           # Application services
│   ├── api/                # HTTP handlers & middleware
│   ├── config/             # Configuration management
│   ├── pkg/                # Shared utilities
│   └── simulator/          # GPU simulation engine
├── web/                    # React frontend
├── deploy/                 # Docker, Helm, Terraform, K8s
├── migrations/             # Database migrations
├── test/                   # Integration & load tests
└── docs/                   # Documentation
```

---

## Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Backend | Go + Gin | Low latency, high concurrency, small binaries |
| Analytics DB | ClickHouse | Column-oriented, 10x faster analytical queries |
| Cache/Queue | Redis | Sub-ms latency, Streams for events |
| Metrics | Prometheus | Pull-based, ecosystem, PromQL |
| Frontend | React + TypeScript | Type safety, component ecosystem |
| Dashboards | Grafana | Standard observability UI |
| Infrastructure | Docker + K8s | Container orchestration |

---

## License

MIT License - see [LICENSE](LICENSE) for details.
