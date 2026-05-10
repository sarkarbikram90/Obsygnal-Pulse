# 🔥 Obsygnal-Pulse

**AI Infrastructure Efficiency Platform** — Detect wasted GPU compute, optimize training workloads, save money.

[![CI](https://github.com/sarkarbikram90/Obsygnal-Pulse/actions/workflows/ci.yml/badge.svg)](https://github.com/sarkarbikram90/Obsygnal-Pulse/actions)

---

## What is Obsygnal-Pulse?

Obsygnal-Pulse is an observability and intelligence platform that monitors GPU clusters running AI/ML workloads, detects inefficiencies (idle GPUs, underutilized memory, stalled training loops), and provides actionable optimization recommendations with confidence scores and estimated cost savings.

### Key Features

- **Real-Time GPU Dashboard** — Cluster overview, GPU utilization heatmaps, time-series metrics charts, live status monitoring
- **6 Detection Rules** — Underutilized GPU, idle GPU, low memory usage, CPU starvation, stalled training, thermal spikes
- **Optimization Recommendations** — Actionable suggestions with confidence scoring (0–100%) and dollar-denominated savings estimates
- **Multi-Channel Alerting** — Severity filtering, deduplication windows, rate limiting, Slack and webhook adapters
- **High-Volume Ingestion** — Validate, deduplicate, buffer, and batch-insert GPU telemetry into ClickHouse
- **GPU Simulator** — 6 realistic behavior profiles for development without GPU hardware

---

## Architecture

```
GPU Telemetry → Ingestion (validate, dedup, buffer) → ClickHouse
                                                         ↓
Detection Engine (periodic eval) ← Query recent metrics ← 
         ↓
Inefficiency → Recommendation Engine → Recommendations
         ↓
Alerting Service → Slack / Webhook
```

- **Backend**: Go 1.26 (Gin), hexagonal architecture, modular monolith
- **Frontend**: React 19, TypeScript, TailwindCSS v4, Recharts, Vite
- **Storage**: ClickHouse (analytics), Redis (cache/dedup), Prometheus (real-time metrics)
- **Infrastructure**: Docker Compose, Kubernetes, Helm, Terraform (GKE)
- **Observability**: Grafana, Prometheus, structured JSON logging (zerolog)

---

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Node.js 22+ (for frontend)
- Make

### 1. Clone & Setup

```bash
git clone https://github.com/sarkarbikram90/Obsygnal-Pulse.git
cd Obsygnal-Pulse
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

Generates realistic GPU telemetry (6 behavior profiles) for development without real GPU hardware.

### 6. Run the Frontend

```bash
cd web
npm install
npm run dev
```

Dashboard available at `http://localhost:5173`

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
| `GET` | `/api/v1/inefficiencies` | List active inefficiencies |
| `GET` | `/api/v1/recommendations` | List optimization recommendations |
| `POST` | `/api/v1/recommendations/:id/apply` | Apply a recommendation |
| `POST` | `/api/v1/recommendations/:id/dismiss` | Dismiss a recommendation |
| `GET` | `/api/v1/alerts/recent` | Recent alert history |

See [docs/api/README.md](docs/api/README.md) for full API reference.

---

## Detection Rules

| Rule | Severity | Trigger |
|------|----------|---------|
| Underutilized GPU | Medium | Avg utilization < 20% |
| Idle GPU | High | Utilization < 2%, 0 processes |
| Low Memory Usage | Medium | Memory utilization < 15% |
| CPU Starvation | High | CPU > 90%, GPU < 50% |
| Stalled Training | Critical | Compute < 5%, memory > 30%, processes > 0 |
| Thermal Spike | Critical | Temperature > 90°C |

---

## Development

```bash
make help          # Show all available commands
make build         # Build all binaries (api-server, simulator, collector, worker)
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint          # Run linters
make fmt           # Format code
make vet           # Run go vet
```

---

## Project Structure

```
├── cmd/                    # Application entry points
│   ├── api-server/         # REST API server
│   ├── collector/          # Prometheus/DCGM telemetry collector
│   ├── worker/             # Background processing (detection pipeline)
│   └── simulator/          # GPU telemetry generator (dev mode)
├── internal/               # Private application code
│   ├── domain/             # Domain models (metrics, detection, recommendation, alerting, auth, cluster)
│   ├── ports/              # Interface definitions (hexagonal architecture)
│   ├── adapters/           # Infrastructure adapters (ClickHouse, Redis, Slack, Webhook, K8s, Prometheus)
│   ├── services/           # Application services (ingestion, detection, recommendation, alerting, analytics, auth)
│   ├── api/                # HTTP handlers, middleware, DTOs, routes
│   ├── config/             # Environment-based configuration
│   ├── pkg/                # Shared utilities (logger, health, errors, validation, tracing)
│   └── simulator/          # GPU simulation engine (6 behavior profiles)
├── web/                    # React + TypeScript frontend (Vite)
├── deploy/                 # Docker, Helm, Terraform, K8s manifests
├── migrations/             # ClickHouse schema migrations
├── test/                   # Integration & load tests
├── docs/                   # ADRs, API reference, architecture, runbooks
└── scripts/                # Development helper scripts
```

---

## Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Backend | Go 1.26 + Gin | Low latency, high concurrency, small binaries |
| Analytics DB | ClickHouse | Column-oriented, 10–100x faster analytical queries |
| Cache/Dedup | Redis | Sub-ms latency, deduplication, connection pooling |
| Metrics | Prometheus | Pull-based, PromQL, DCGM Exporter compatible |
| Frontend | React 19 + TypeScript + Vite | Type safety, fast HMR, modern tooling |
| Charts | Recharts | React-native composable chart library |
| Styling | TailwindCSS v4 | Utility-first, co-located, dark theme |
| Dashboards | Grafana | Standard observability UI |
| Infrastructure | Docker + K8s + Helm + Terraform | Container orchestration, IaC |
| CI/CD | GitHub Actions | Lint → Test → Build → Docker → Frontend pipeline |

---

## Test Suite

```
56 unit tests across 7 packages — all passing
 3 integration tests (gated by INTEGRATION_TEST=1)
 2 load tests (gated by LOAD_TEST=1)
```

Run the full suite:

```bash
make test
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.
