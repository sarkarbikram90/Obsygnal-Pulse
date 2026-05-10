package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/simulator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.Environment)

	simCfg := simulator.DefaultConfig()
	sim := simulator.New(simCfg, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start a simple HTTP server to expose simulated metrics for health checks
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	mux.HandleFunc("/metrics/latest", func(w http.ResponseWriter, r *http.Request) {
		batch := sim.GenerateTick(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(batch)
	})

	port := 8085
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Info().Int("port", port).Msg("simulator HTTP server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// For MVP, publish directly to stdout (will be replaced with Redis publisher)
	publisher := func(ctx context.Context, batch *metrics.MetricBatch) error {
		log.Info().
			Int("gpu_metrics", len(batch.GPUMetrics)).
			Int("node_metrics", len(batch.NodeMetrics)).
			Str("source", batch.Source).
			Msg("telemetry batch generated")

		// In development, also write to a local endpoint for the collector
		return nil
	}

	// Run simulator in background
	go func() {
		if err := sim.Run(ctx, publisher); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("simulator error")
		}
	}()

	// Wait for shutdown signal
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down simulator")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}
