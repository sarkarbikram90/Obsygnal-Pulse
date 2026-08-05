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
	memoryadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/memory"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/nats"
	redisadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/redis"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	detectiondomain "github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/recommendation"
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

	var metricsRepo ports.MetricsRepository
	var ineffRepo ports.InefficiencyRepository
	var recRepo ports.RecommendationRepository

	// Initialize infrastructure adapters with fallback for standalone/dev environments
	chAdapter, err := clickhouse.NewAdapter(cfg.ClickHouse, log)
	if err == nil {
		defer chAdapter.Close()
		metricsRepo = chAdapter
		ineffRepo = clickhouse.NewInefficiencyAdapter(chAdapter.Conn(), log)
		recRepo = clickhouse.NewRecommendationAdapter(chAdapter.Conn(), log)
		log.Info().Msg("worker connected to ClickHouse OLAP")
	} else {
		log.Warn().Err(err).Msg("ClickHouse unavailable — worker initializing stateful in-memory repository")
		memStore := memoryadapter.NewRepositoryStore(log)
		metricsRepo = memStore
		ineffRepo = memStore.InefficiencyRepo()
		recRepo = memStore.RecommendationRepo()
	}

	var cacheStore ports.CacheStore
	rAdapter, err := redisadapter.NewAdapter(cfg.Redis, log)
	if err == nil {
		defer rAdapter.Close()
		cacheStore = rAdapter
	} else {
		log.Warn().Err(err).Msg("Redis unavailable — worker running without cache layer")
	}

	// Initialize ingestion service
	ingestionSvc := ingestion.NewService(
		metricsRepo,
		cacheStore,
		log,
		cfg.ClickHouse.BatchSize,
		cfg.ClickHouse.FlushInterval,
	)

	// Start background flush loop for the batcher
	ingestionSvc.StartPeriodicFlush(ctx)

	// Subscribe to NATS telemetry stream if available
	natsAdapter, err := nats.NewAdapter(cfg.Nats, log)
	if err == nil {
		defer natsAdapter.Close()
		log.Info().Msg("subscribing to NATS telemetry stream...")
		if err := natsAdapter.SubscribeIngester(ctx, func(msgCtx context.Context, batch *metrics.MetricBatch) error {
			return ingestionSvc.ProcessBatch(msgCtx, batch)
		}); err != nil {
			log.Warn().Err(err).Msg("failed to subscribe to NATS stream")
		}
	} else {
		log.Warn().Err(err).Msg("NATS unavailable — worker running in standalone evaluation mode")
	}

	detectionSvc := detection.NewService(metricsRepo, ineffRepo, cacheStore, log, cfg.Detection)
	recommendationSvc := recommendation.NewService(recRepo, cacheStore, log)

	// Wire detection → recommendation pipeline
	detectionSvc.OnDetection(func(ctx context.Context, ineff detectiondomain.Inefficiency) {
		recs, err := recommendationSvc.GenerateFromInefficiency(ctx, ineff)
		if err != nil {
			log.Error().Err(err).Str("inefficiency_id", ineff.ID).Msg("failed to generate recommendations")
			return
		}
		if len(recs) > 0 {
			log.Info().Int("count", len(recs)).Str("inefficiency_id", ineff.ID).Msg("generated recommendations")
		}
	})

	go func() {
		if err := detectionSvc.Run(ctx); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("detection engine stopped with error")
		}
	}()

	log.Info().Msg("worker running — waiting for shutdown signal")

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down worker")

	cancel()

	// Final flush
	ingestionSvc.Flush(context.Background())
}
