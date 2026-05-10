# API Layer (`internal/api/`)

## Overview
The `api` directory is the primary entry point for HTTP traffic (Primary Adapter). It manages the REST API surface for the frontend dashboard and programmatic data ingestion.

## Responsibilities
- Parsing and validating incoming HTTP requests and JSON payloads.
- Routing URLs to the correct handler functions.
- Managing HTTP middleware (Authentication, Rate Limiting, CORS, Security Headers).
- Translating core domain errors into appropriate HTTP status codes (e.g., 400 Bad Request, 404 Not Found, 500 Internal Server Error).
- Serializing domain models into JSON responses.

## Architecture Role
This is a Primary Adapter in the Hexagonal Architecture. It drives the application by invoking methods on the `internal/services/` layer. It must NEVER contain business logic or make direct database queries.

## Key Components
- `handlers/`: Contains the actual HTTP handler functions. These are kept extremely thin.
- `routes/`: Centralizes the URL topology and dependency injection wiring for the HTTP server.
- `middleware/`: Reusable HTTP interception logic (e.g., `ratelimit.go`, `auth.go`, `logger.go`).
- `dto/`: Data Transfer Objects. Defines the exact JSON shapes expected in requests and returned in responses.

## Dependencies
- **Internal**: `internal/services`, `internal/domain`, `internal/pkg/logger`.
- **External**: `github.com/gin-gonic/gin` (HTTP Framework).

## Data Flow
```text
HTTP Request → Middleware (Auth, RateLimit) → Routes → Handler (Parsing) → Service → Handler (Response) → HTTP Response
```

## Operational Considerations
- **Scaling**: Handlers are stateless. Throughput is primarily limited by the backend infrastructure (ClickHouse/Redis) and network bandwidth.
- **Reliability**: Panic recovery middleware ensures that a panic in one request does not crash the entire server process.

## Security Considerations
- **Authentication**: JWT Bearer token validation is enforced via `auth.go`.
- **Rate Limiting**: Sliding window rate limiting via Redis (`ratelimit.go`) protects against L7 DDoS and noisy neighbors.
- **Payload Validation**: All incoming requests must be strictly validated against struct tags in the `dto` package to prevent injection attacks.

## Observability
- **Metrics**: Endpoints should expose `http_request_duration_seconds` to Prometheus.
- **Logging**: The `logger.go` middleware logs every request's latency, status code, and IP address.
- **Tracing**: Request IDs are generated for tracing logs across the request lifecycle.

## Testing
- **Unit Tests**: Handlers are tested using `httptest.NewRecorder()` and mocked services.
- **API Contracts**: Tests must verify the exact JSON shape to ensure backward compatibility.

## Future Improvements
- Implement OpenAPI/Swagger generation directly from the Go code to keep API documentation in sync.
- Migrate from simple JWT token verification to a full JWKS (JSON Web Key Set) implementation for integration with external Identity Providers (OIDC).
