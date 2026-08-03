// Package main implements the edge telemetry collector daemon.
// The collector is deployed as a DaemonSet or agent on GPU nodes.
// It queries local Prometheus / DCGM exporters and pushes metrics to the upstream API.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/prometheus"
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
		Msg("starting Obsygnal-Pulse edge collector")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	promCollector := prometheus.NewAdapter(cfg.Prometheus.URL, 10*time.Second)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	apiIngestURL := fmt.Sprintf("http://%s:%d/api/v1/metrics/ingest", cfg.Server.Host, cfg.Server.Port)

	ticker := time.NewTicker(cfg.Collector.ScrapeInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				batch, err := promCollector.Collect(ctx)
				if err != nil {
					log.Warn().Err(err).Msg("failed to scrape telemetry batch")
					continue
				}

				if len(batch.GPUMetrics) == 0 {
					continue
				}

				payload, err := json.Marshal(batch)
				if err != nil {
					log.Error().Err(err).Msg("failed to marshal metric batch")
					continue
				}

				req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiIngestURL, bytes.NewBuffer(payload))
				if err != nil {
					log.Error().Err(err).Msg("failed to build ingest request")
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer obsygnal_sec_production_agent")

				resp, err := httpClient.Do(req)
				if err != nil {
					log.Warn().Err(err).Str("url", apiIngestURL).Msg("failed to push telemetry batch to API server")
					continue
				}
				resp.Body.Close()

				log.Info().
					Int("gpu_metrics", len(batch.GPUMetrics)).
					Int("status_code", resp.StatusCode).
					Msg("telemetry batch pushed successfully")
			}
		}
	}()

	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down collector daemon")
	cancel()
}
