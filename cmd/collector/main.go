// Package main implements the edge telemetry collector daemon.
// The collector is intended to be deployed as a DaemonSet on Kubernetes GPU nodes.
// It directly interfaces with the NVIDIA Management Library (NVML) to scrape hardware metrics
// and securely streams the batched payloads to the upstream api-server for aggregation.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/nats"
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

	natsAdapter, err := nats.NewAdapter(cfg.Nats, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to NATS JetStream")
	}
	defer natsAdapter.Close()

	// TODO: Initialize Prometheus client and start scrape loop
	// collector := prometheus.NewCollector(cfg.Prometheus)
	// loop := collector.RunScrapeLoop(ctx, cfg.Collector.ScrapeInterval, natsAdapter.PublishBatch)

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
