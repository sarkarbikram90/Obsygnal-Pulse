// Package main implements the background worker service.
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

	"github.com/obsygnal/obsygnal-pulse/internal/config"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
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

	// TODO: Initialize services with real adapters and start processing loops
	// metricsRepo := clickhouse.NewAdapter(cfg.ClickHouse)
	// cache := redis.NewAdapter(cfg.Redis)
	// detectionSvc := detection.NewService(metricsRepo, ineffRepo, cache, log, cfg.Detection)
	// recommendationSvc := recommendation.NewService(recRepo, cache, log)
	// alertingSvc := alerting.NewService(alertRepo, log, cfg.Alerting)
	//
	// // Wire detection → recommendation → alerting pipeline
	// detectionSvc.OnDetection(func(ctx context.Context, ineff detection.Inefficiency) {
	//     recommendationSvc.GenerateFromInefficiency(ctx, ineff)
	//     alertingSvc.HandleInefficiency(ctx, ineff)
	// })
	//
	// go detectionSvc.Run(ctx)

	log.Info().Msg("worker running — waiting for shutdown signal")
	_ = ctx // Will be used when services are wired

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down worker")
	cancel()
}
