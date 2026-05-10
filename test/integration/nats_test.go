package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/nats"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// TestNatsIngestionPipeline tests the full publish/subscribe lifecycle using JetStream.
// Requires NATS to be running on localhost:4222
func TestNatsIngestionPipeline(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping NATS integration test. Set INTEGRATION_TEST=1 to run.")
	}

	log := logger.New("development")
	cfg := config.NatsConfig{
		URL:        "nats://localhost:4222",
		StreamName: "telemetry_test_stream",
		MaxRetries: 3,
		AckWait:    2 * time.Second,
	}

	adapter, err := nats.NewAdapter(cfg, log)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer adapter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Subscribe to the stream
	receivedCh := make(chan *metrics.MetricBatch, 1)
	err = adapter.SubscribeIngester(ctx, func(msgCtx context.Context, batch *metrics.MetricBatch) error {
		receivedCh <- batch
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 2. Publish a test batch
	testBatch := &metrics.MetricBatch{
		Source: "integration_test",
		GPUMetrics: []metrics.GPUMetric{
			{GPUID: "gpu-test-1", Utilization: 85.0},
		},
	}

	err = adapter.PublishBatch(ctx, testBatch)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// 3. Verify receipt
	select {
	case <-ctx.Done():
		t.Fatal("Timeout waiting for message")
	case received := <-receivedCh:
		if received.Source != testBatch.Source {
			t.Errorf("Expected source %s, got %s", testBatch.Source, received.Source)
		}
		if len(received.GPUMetrics) != 1 || received.GPUMetrics[0].GPUID != "gpu-test-1" {
			t.Errorf("Received incorrect payload")
		}
	}
}

// TestNatsChaosDLQ tests that a poison message is correctly routed to the DLQ.
func TestNatsChaosDLQ(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping NATS chaos test. Set INTEGRATION_TEST=1 to run.")
	}

	log := logger.New("development")
	cfg := config.NatsConfig{
		URL:        "nats://localhost:4222",
		StreamName: "telemetry_chaos_stream",
		MaxRetries: 1, // Fail fast for tests
		AckWait:    1 * time.Second,
	}

	adapter, err := nats.NewAdapter(cfg, log)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer adapter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fail the processing intentionally to trigger DLQ
	errCount := 0
	err = adapter.SubscribeIngester(ctx, func(msgCtx context.Context, batch *metrics.MetricBatch) error {
		errCount++
		return context.DeadlineExceeded // simulate a crash or timeout
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	testBatch := &metrics.MetricBatch{Source: "poison_pill"}
	err = adapter.PublishBatch(ctx, testBatch)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for retry loop to finish
	time.Sleep(3 * time.Second)

	if errCount < 1 {
		t.Errorf("Expected poison pill to be attempted, got %d attempts", errCount)
	}
}
