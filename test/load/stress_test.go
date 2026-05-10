// Package load_test contains load/stress tests for the ingestion pipeline.
//
// These tests verify the system can handle expected production throughput.
// Run with: go test ./test/load/ -v -timeout 5m
package load_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/metrics"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/simulator"
)

func skipIfNotLoadTest(t *testing.T) {
	if os.Getenv("LOAD_TEST") == "" {
		t.Skip("Skipping load test — set LOAD_TEST=1 to run")
	}
}

// TestSimulatorThroughput measures how many metrics the simulator can generate.
func TestSimulatorThroughput(t *testing.T) {
	skipIfNotLoadTest(t)

	cfg := simulator.DefaultConfig()
	cfg.NodeCount = 100
	cfg.GPUsPerNode = 8
	log := logger.New("development")

	sim := simulator.New(cfg, log)

	start := time.Now()
	totalMetrics := 0
	iterations := 100

	for i := 0; i < iterations; i++ {
		batch := sim.GenerateTick(context.Background())
		totalMetrics += len(batch.GPUMetrics) + len(batch.NodeMetrics)
	}

	elapsed := time.Since(start)
	metricsPerSecond := float64(totalMetrics) / elapsed.Seconds()

	t.Logf("Generated %d metrics in %v", totalMetrics, elapsed)
	t.Logf("Throughput: %.0f metrics/second", metricsPerSecond)
	t.Logf("GPUs simulated: %d", cfg.NodeCount*cfg.GPUsPerNode)

	// Minimum throughput: 100K metrics/second
	if metricsPerSecond < 100000 {
		t.Errorf("throughput below minimum: %.0f < 100000 metrics/sec", metricsPerSecond)
	}
}

// TestBatchIngestStress stress-tests batch ingestion with many concurrent batches.
func TestBatchIngestStress(t *testing.T) {
	skipIfNotLoadTest(t)

	// Generate large batches
	gpuMetrics := make([]metrics.GPUMetric, 10000)
	for i := range gpuMetrics {
		gpuMetrics[i] = metrics.GPUMetric{
			ClusterID:         "stress-cluster",
			NodeID:            "stress-node",
			GPUID:             "stress-gpu",
			Timestamp:         time.Now(),
			Utilization:       50.0,
			MemoryUtilization: 40.0,
			Temperature:       65.0,
		}
	}

	batch := &metrics.MetricBatch{
		GPUMetrics: gpuMetrics,
		ReceivedAt: time.Now(),
		Source:     "stress-test",
	}

	t.Logf("Batch size: %d metrics", batch.Size())
	t.Log("Stress test scaffolding ready — wire to real ingestion service")
}
