// Package main defines the primary entrypoint for the api-server.
// The api-server provides the public-facing REST API for Obsygnal-Pulse.
// It is responsible for ingesting high-throughput GPU telemetry from nodes,
// serving aggregated cluster health dashboards, and exposing optimization recommendations.
// It acts as the Primary Adapter, safely wiring HTTP routes to internal domain services.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/clickhouse"
	redisadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/redis"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/routes"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

const version = "0.1.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Server.Environment)
	log.Info().
		Str("version", version).
		Str("environment", cfg.Server.Environment).
		Msg("starting Obsygnal-Pulse API server")

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

	// Initialize health checks
	healthRegistry := health.NewRegistry(version)
	healthRegistry.Register("clickhouse", func(ctx context.Context) error {
		return chAdapter.Ping(ctx)
	})
	healthRegistry.Register("redis", func(ctx context.Context) error {
		return redisAdapter.Ping(ctx)
	})

	// Initialize services
	ingestionSvc := ingestion.NewService(
		chAdapter,
		redisAdapter,
		log,
		cfg.ClickHouse.BatchSize,
		cfg.ClickHouse.FlushInterval,
	)

	// Start periodic flush
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ingestionSvc.StartPeriodicFlush(ctx)

	// Setup Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	authService := authsvc.NewService("obsygnal-pulse-secret-mvp", time.Hour, log)

	routes.Register(router, routes.Deps{
		Log:          log,
		Registry:     healthRegistry,
		MetricsRepo:  chAdapter,
		IngestionSvc: ingestionSvc,
		RedisClient:  redisAdapter.Client(),
		AuthService:  authService,
	})

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	cancel() // Stop background workers

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}

	// Final flush
	ingestionSvc.Flush(context.Background())

	log.Info().Msg("server stopped")
}
