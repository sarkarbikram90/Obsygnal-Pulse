// Package main implements the telemetry collector service.
//
// The collector scrapes GPU metrics from Prometheus (which in turn
// scrapes DCGM Exporter) and publishes them to Redis Streams for
// the ingestion service to consume.
//
// In production, this replaces the simulator.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.Environment)
	log.Info().
		Str("prometheus_url", cfg.Prometheus.URL).
		Dur("scrape_interval", cfg.Collector.ScrapeInterval).
		Msg("starting telemetry collector")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Initialize Prometheus client, Redis publisher, and start scrape loop
	// collector := prometheus.NewCollector(cfg.Prometheus)
	// publisher := redis.NewPublisher(cfg.Redis)
	// loop := collector.RunScrapeLoop(ctx, cfg.Collector.ScrapeInterval, publisher)

	ticker := time.NewTicker(cfg.Collector.ScrapeInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Debug().Msg("collector tick (not yet implemented)")
			}
		}
	}()

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down collector")
	cancel()
}
