// Package main implements the asynchronous background worker service for Obsygnal-Pulse.
//
// The worker processes asynchronous tasks such as:
// - Running the detection engine evaluation loop
// - Generating recommendations from detected inefficiencies
// - Dispatching alerts to notification channels
// - Cleaning up expired recommendations and resolved inefficiencies
// - Running periodic analytics aggregation jobs
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/clickhouse"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/nats"
	redisadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/redis"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.Environment)
	log.Info().Msg("starting background worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Initialize infrastructure adapters
	chAdapter, err := clickhouse.NewAdapter(cfg.ClickHouse, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to ClickHouse")
	}
	defer chAdapter.Close()

	redisAdapter, err := redisadapter.NewAdapter(cfg.Redis, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer redisAdapter.Close()

	natsAdapter, err := nats.NewAdapter(cfg.Nats, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to NATS JetStream")
	}
	defer natsAdapter.Close()

	// Initialize ingestion service
	ingestionSvc := ingestion.NewService(
		chAdapter,
		redisAdapter,
		log,
		cfg.ClickHouse.BatchSize,
		cfg.ClickHouse.FlushInterval,
	)

	// Start background flush loop for the batcher
	ingestionSvc.StartPeriodicFlush(ctx)

	// Subscribe to NATS telemetry stream
	log.Info().Msg("subscribing to telemetry stream...")
	err = natsAdapter.SubscribeIngester(ctx, func(msgCtx context.Context, batch *metrics.MetricBatch) error {
		return ingestionSvc.ProcessBatch(msgCtx, batch)
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start NATS consumer")
	}

	// TODO: Initialize detection/recommendation engines
	// detectionSvc := detection.NewService(...)
	// go detectionSvc.Run(ctx)

	log.Info().Msg("worker running — waiting for shutdown signal")

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down worker")

	cancel()

	// Final flush
	ingestionSvc.Flush(context.Background())
}
