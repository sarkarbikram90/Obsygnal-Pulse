#!/usr/bin/env bash
# =============================================================================
# dev-setup.sh — Bootstraps the complete local Kubernetes-equivalent development environment.
# This script installs Go dependencies, provisions the ClickHouse+Redis+Prometheus topology
# via Docker Compose, waits for TCP readiness, and automatically applies SQL migrations.
# =============================================================================
set -euo pipefail

echo "🔥 Obsygnal-Pulse Development Setup"
echo "==================================="

# Check prerequisites
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "❌ $1 is required but not installed."
        exit 1
    fi
    echo "✅ $1 found"
}

echo ""
echo "Checking prerequisites..."
check_command go
check_command docker
check_command make

# Copy env file if not exists
if [ ! -f .env ]; then
    echo ""
    echo "📋 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ .env created — edit it with your local settings"
else
    echo "✅ .env already exists"
fi

# Install Go dependencies
echo ""
echo "📦 Installing Go dependencies..."
go mod download
echo "✅ Go dependencies installed"

# Start infrastructure
echo ""
echo "🐳 Starting infrastructure (ClickHouse, Redis, Prometheus, Grafana)..."
docker compose up -d
echo "✅ Infrastructure started"

# Wait for ClickHouse
echo ""
echo "⏳ Waiting for ClickHouse to be ready..."
for i in {1..30}; do
    if docker compose exec -T clickhouse clickhouse-client --query "SELECT 1" &> /dev/null; then
        echo "✅ ClickHouse is ready"
        break
    fi
    sleep 1
done

# Run migrations
echo ""
echo "🗃️  Running database migrations..."
make migrate-docker 2>/dev/null || echo "⚠️  Migration skipped (run manually with: make migrate-docker)"

# Build binaries
echo ""
echo "🔨 Building binaries..."
make build
echo "✅ Binaries built"

# Run tests
echo ""
echo "🧪 Running tests..."
make test
echo ""

echo "==================================="
echo "🎉 Setup complete!"
echo ""
echo "Quick start:"
echo "  make dev            — Start the API server"
echo "  make dev-simulator  — Start the GPU simulator"
echo "  make docker-up      — Start infrastructure"
echo "  make test           — Run tests"
echo ""
echo "Dashboards:"
echo "  Grafana:     http://localhost:3000 (admin / obsygnal)"
echo "  Prometheus:  http://localhost:9090"
echo "  ClickHouse:  http://localhost:8123"
echo "==================================="
