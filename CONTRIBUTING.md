# Contributing to Obsygnal-Pulse

Thank you for your interest in contributing to **Obsygnal-Pulse**! We are building the open-source standard for AI infrastructure observability, deep GPU telemetry, and compute cost optimization.

This document provides guidelines and workflows for contributing to the project.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Architecture & Mental Model](#architecture--mental-model)
3. [Development Environment Setup](#development-environment-setup)
4. [Development Workflow](#development-workflow)
5. [Coding Standards & Conventions](#coding-standards--conventions)
6. [Testing Standards](#testing-standards)
7. [Commit Message Guidelines](#commit-message-guidelines)
8. [Pull Request Process](#pull-request-process)
9. [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)

---

## Code of Conduct

All contributors and maintainers are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md). Please report unacceptable behavior to [security@obsygnal.io](mailto:security@obsygnal.io).

---

## Architecture & Mental Model

Obsygnal-Pulse is structured as a modular monolith adhering strictly to **Hexagonal Architecture (Ports & Adapters)**:

```
┌─────────────────────────────────────────────────────────┐
│                      cmd/ (Entry Points)                │
│       api-server  |  collector  |  worker  | simulator  │
└────────────────────────────┬────────────────────────────┘
                             │ (Wires dependencies)
                             ▼
┌─────────────────────────────────────────────────────────┐
│              internal/api/ (Primary Adapters)           │
│           Handlers  |  Middleware  |  Routes  |  DTOs    │
└────────────────────────────┬────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────┐
│             internal/services/ (Core Orchestration)     │
│   Ingestion | Detection | Recommendation | Auth | Alerts│
└──────────────┬───────────────────────────┬──────────────┘
               │                           │
               ▼                           ▼
┌──────────────────────────────┐ ┌────────────────────────┐
│   internal/domain/ (Pure)    │ │ internal/ports/ (Pure) │
│ Metrics, Detection, Auth, ROI│ │ Repository Interfaces  │
└──────────────────────────────┘ └───────────┬────────────┘
                                             │ (Implemented by)
                                             ▼
┌─────────────────────────────────────────────────────────┐
│             internal/adapters/ (Secondary Adapters)     │
│     ClickHouse | Redis | NATS | Prometheus | Memory     │
└─────────────────────────────────────────────────────────┘
```

### Key Architectural Rules:
- **`internal/domain`** contains pure business logic and models. It must **never** import other internal packages (except `internal/pkg/errors`) or external infrastructure libraries.
- **`internal/ports`** defines interfaces for databases, caches, and notification services.
- **`internal/services`** depends only on `domain` models and `ports` interfaces. It never directly imports `adapters`.
- **`internal/adapters`** implements the port interfaces for specific infrastructure (ClickHouse, Redis, Prometheus, Slack).
- **`cmd/`** is the only place where concrete adapters are instantiated and injected into services (the composition root).
- **Zero-Dependency Fallback**: All components must gracefully fall back to thread-safe in-memory implementations if external infrastructure (ClickHouse, Redis) is unavailable.

---

## Development Environment Setup

### Prerequisites
- **Go**: `1.24+` ([Download Go](https://golang.org/dl/))
- **Node.js**: `v20+` and `npm` (for the React dashboard)
- **Docker & Docker Compose**: (Optional, for local ClickHouse, Redis, and NATS)
- **golangci-lint**: `v1.60+` (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### 1. Clone the Repository
```bash
git clone https://github.com/sarkarbikram90/Obsygnal-Pulse.git
cd Obsygnal-Pulse
```

### 2. Environment Configuration
```bash
cp .env.example .env
```

### 3. Start Local Infrastructure (Optional)
```bash
make docker-up
```
This spins up:
- **ClickHouse**: `http://localhost:8123` (Native port: `9000`)
- **Redis**: `localhost:6379`
- **NATS JetStream**: `localhost:4222`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (`admin`/`obsygnal`)

### 4. Run the Backend API Server
```bash
go run ./cmd/api-server/main.go
```
*The API server will automatically initialize in-memory fallbacks if ClickHouse/Redis are not running.*

### 5. Run the React Web Dashboard
```bash
cd web
npm install
npm run dev
```
Open [http://localhost:5173](http://localhost:5173) in your browser.

---

## Development Workflow

### Useful Makefile Targets
| Command | Description |
| :--- | :--- |
| `make build` | Build all Go binaries into `bin/` |
| `make test` | Run all unit, integration, and E2E tests |
| `make test-coverage` | Run test suite and generate HTML coverage report |
| `make lint` | Run `golangci-lint` static analysis |
| `make fmt` | Format Go code with `gofmt` and `goimports` |
| `make vet` | Run `go vet ./...` |
| `make dev` | Run API server in development mode |
| `make dev-simulator` | Run GPU telemetry simulator |
| `make docker-up` | Start ClickHouse, Redis, NATS, Prometheus, Grafana |
| `make docker-down` | Stop local infrastructure containers |
| `make clean` | Remove compiled binaries and test coverage artifacts |

---

## Coding Standards & Conventions

### Go Guidelines
- Follow standard Go naming conventions ([Effective Go](https://golang.org/doc/effective_go)).
- **Errors**: Always wrap errors with context using `fmt.Errorf("action failed: %w", err)` or `internal/pkg/errors`.
- **Concurrency**: Protect shared state with `sync.RWMutex` or `sync.Mutex`. Verify thread safety using `go test -race ./...`.
- **Logging**: Use structured logging via `internal/pkg/logger` (`zerolog`). Do not use `fmt.Println` or standard `log` in production code.
- **Context**: Pass `context.Context` as the first argument in all network, database, and long-running operations.

### Frontend Guidelines
- Use **TypeScript** with strict type safety (avoid `any`).
- Use **TailwindCSS** utility classes for styling.
- Use **Lucide React** for icons.
- Prefer React Query (`@tanstack/react-query`) for server state management and caching.

---

## Testing Standards

Quality and reliability are paramount for infrastructure software.

1. **Unit Tests**: Place unit tests in `*_test.go` files alongside the code being tested.
2. **Table-Driven Tests**: Prefer table-driven testing for multi-case evaluations (e.g. detection rules).
3. **Mocking**: Use interface doubles or lightweight memory stores rather than heavy mocking frameworks.
4. **Coverage Target**: Maintain statement coverage of **>= 90%** across `internal/domain`, `internal/services`, and `internal/api`.

Run test validation locally:
```bash
go test -v -race -cover ./internal/...
```

---

## Commit Message Guidelines

We enforce the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

### Allowed Types:
- `feat`: A new feature or detection rule
- `fix`: A bug fix
- `docs`: Documentation changes
- `perf`: Performance improvements
- `refactor`: Code refactoring without behavioral changes
- `test`: Adding or correcting tests
- `chore`: Build scripts, CI workflow, or dependency updates

### Example:
```
feat(detection): add NCCL distributed communication straggler rule

Detects GPU ranks stalled in NCCL all-reduce collectives exceeding
the 15-second latency threshold.

Closes #42
```

---

## Pull Request Process

1. **Fork the repository** and create a feature branch (`git checkout -b feat/my-new-feature`).
2. **Implement your changes** and add comprehensive unit tests.
3. **Run local validation**:
   ```bash
   make fmt
   make vet
   make test
   ```
4. **Commit your changes** following Conventional Commits.
5. **Push your branch** to your fork and submit a Pull Request to `main`.
6. Ensure all CI checks (linting, race detector, static binaries, Trivy scanning, frontend build) pass.
7. Address reviewer feedback promptly.

---

## Developer Certificate of Origin (DCO)

By submitting a Pull Request, you certify that you have the right to submit the code under the project's [MIT License](LICENSE), in accordance with the [Developer Certificate of Origin v1.1](https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

To sign off on your commits, use the `-s` flag with `git commit`:
```bash
git commit -s -m "feat(collector): add NVLink bandwidth metric scraper"
```
