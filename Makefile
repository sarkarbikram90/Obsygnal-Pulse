.PHONY: help build test lint run dev docker-up docker-down migrate clean

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================
# Build
# ============================================================
build: ## Build all services
	go build -o bin/api-server ./cmd/api-server/
	go build -o bin/collector ./cmd/collector/
	go build -o bin/worker ./cmd/worker/
	go build -o bin/simulator ./cmd/simulator/

build-api: ## Build API server only
	go build -o bin/api-server ./cmd/api-server/

build-simulator: ## Build simulator only
	go build -o bin/simulator ./cmd/simulator/

# ============================================================
# Test
# ============================================================
test: ## Run all tests
	go test -v -race -count=1 ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-unit: ## Run unit tests only
	go test -v -race -short ./internal/...

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./test/integration/...

# ============================================================
# Lint
# ============================================================
lint: ## Run linters
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w -s .
	goimports -w .

vet: ## Run go vet
	go vet ./...

# ============================================================
# Development
# ============================================================
dev: ## Run API server in development mode
	ENVIRONMENT=development go run ./cmd/api-server/

dev-simulator: ## Run GPU simulator
	ENVIRONMENT=development go run ./cmd/simulator/

# ============================================================
# Docker
# ============================================================
docker-up: ## Start infrastructure services
	docker compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo "ClickHouse: http://localhost:8123"
	@echo "Redis:      localhost:6379"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana:    http://localhost:3000 (admin/obsygnal)"

docker-down: ## Stop infrastructure services
	docker compose down

docker-clean: ## Stop and remove all data
	docker compose down -v

docker-build: ## Build Docker images
	docker build -f deploy/docker/api-server.Dockerfile --build-arg SERVICE=api-server -t obsygnal-pulse/api-server:latest .
	docker build -f deploy/docker/api-server.Dockerfile --build-arg SERVICE=simulator -t obsygnal-pulse/simulator:latest .

# ============================================================
# Database
# ============================================================
migrate: ## Run ClickHouse migrations
	@echo "Applying ClickHouse migrations..."
	@for f in migrations/clickhouse/*.sql; do \
		echo "Applying $$f..."; \
		clickhouse-client --host localhost --password obsygnal_dev --multiquery < "$$f"; \
	done
	@echo "Migrations complete."

migrate-docker: ## Run migrations via Docker
	@for f in migrations/clickhouse/*.sql; do \
		echo "Applying $$f..."; \
		docker exec -i obsygnal-clickhouse clickhouse-client --password obsygnal_dev --multiquery < "$$f"; \
	done

# ============================================================
# Dependencies
# ============================================================
deps: ## Download dependencies
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# ============================================================
# Clean
# ============================================================
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# ============================================================
# Setup
# ============================================================
setup: ## Initial project setup
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Copying .env..."
	@cp -n .env.example .env 2>/dev/null || true
	@echo "Setup complete. Run 'make docker-up' to start infrastructure."
