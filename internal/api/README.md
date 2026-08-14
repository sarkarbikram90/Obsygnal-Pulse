# API Layer (`internal/api/`)

## Overview
The `api` package acts as the Primary Adapter (driving adapter) in Obsygnal-Pulse's Hexagonal Architecture. It manages the REST API interface, routes incoming HTTP requests, enforces authentication/RBAC, and serializes domain models into JSON responses.

---

## Directory Structure

```
internal/api/
├── dto/            # Data Transfer Objects (request/response schemas with validation tags)
├── handlers/       # HTTP handler functions (thin orchestration calling services)
│   ├── handlers.go       # Core telemetry, inefficiencies, and recommendations handlers
│   ├── saas_handlers.go  # Multi-tenant auth, API keys, billing usage, audit log handlers
│   └── prometheus.go     # Operational /metrics endpoint handler
├── middleware/     # Reusable HTTP interception logic
│   ├── auth.go           # JWT token validation, TenantScope, and RequirePermission RBAC
│   ├── ratelimit.go      # Sliding-window Redis rate limiter with in-memory fallback
│   └── middleware.go     # CORS, RequestID tracing, Logger, and Panic Recovery
└── routes/         # Central route registration and dependency injection wiring
```

---

## Request Lifecycle

```text
HTTP Request
     │
     ▼
[Middleware: Panic Recovery & Tracing]
     │
     ▼
[Middleware: CORS & Security Headers]
     │
     ▼
[Middleware: Redis Rate Limiter]
     │
     ▼
[Middleware: JWT / Machine API Key Auth & RBAC]
     │
     ▼
[Handler: Parse & Validate DTO Request]
     │
     ▼
[Service Layer: Business Orchestration]
     │
     ▼
[Handler: Format DTO Response & Status Code]
     │
     ▼
HTTP Response
```

---

## Testing Handlers

Handlers are tested using `net/http/httptest` and memory repositories without requiring running external servers:
```bash
go test -v ./internal/api/...
```
