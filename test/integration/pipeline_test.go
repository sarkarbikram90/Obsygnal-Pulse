// Package integration_test contains end-to-end integration tests.
//
// These tests require running infrastructure (ClickHouse, Redis).
// They are skipped in CI unless the INTEGRATION_TEST env var is set.
//
// Run with: INTEGRATION_TEST=1 go test ./test/integration/ -v
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	detectionSvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/detection"
	ingestionSvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test — set INTEGRATION_TEST=1 to run")
	}
}

// TestEndToEndPipeline verifies the full flow:
// Ingest metrics → Detect inefficiencies → Generate recommendations
func TestEndToEndPipeline(t *testing.T) {
	skipIfNotIntegration(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	log := logger.New(cfg.Server.Environment)
	ctx := context.Background()

	// TODO: Initialize real ClickHouse + Redis adapters
	// metricsRepo := clickhouse.NewAdapter(cfg.ClickHouse)
	// cache := redis.NewAdapter(cfg.Redis)

	_ = log
	_ = ctx

	t.Log("integration test scaffolding ready — implement with real adapters")
}

// TestIngestionThroughput measures ingestion throughput.
func TestIngestionThroughput(t *testing.T) {
	skipIfNotIntegration(t)

	log := logger.New("development")
	_ = ingestionSvc.NewService(nil, nil, log, 10000, 5*time.Second)

	t.Log("throughput test scaffolding ready")
}

// TestDetectionAccuracy verifies detection rules produce correct results.
func TestDetectionAccuracy(t *testing.T) {
	skipIfNotIntegration(t)

	cfg := config.DetectionConfig{
		EvaluationInterval:         60 * time.Second,
		GPUUtilizationThreshold:    20.0,
		MemoryUtilizationThreshold: 15.0,
	}

	_ = detectionSvc.NewService(nil, nil, nil, logger.New("development"), cfg)
	_ = metrics.GPUMetric{}

	t.Log("detection accuracy test scaffolding ready")
}
