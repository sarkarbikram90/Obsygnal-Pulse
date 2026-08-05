// Package main defines the primary entrypoint for the api-server.
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
	"github.com/redis/go-redis/v9"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/clickhouse"
	memoryadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/memory"
	redisadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/redis"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/routes"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
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

	// Initialize health checks
	healthRegistry := health.NewRegistry(version)

	var metricsRepo ports.MetricsRepository
	var ineffRepo ports.InefficiencyRepository
	var recRepo ports.RecommendationRepository
	var alertRepo ports.AlertRepository

	var usageRepo ports.UsageRepository
	var authRepo ports.AuthRepository
	var apiKeyRepo ports.APIKeyRepository
	var auditRepo ports.AuditLogRepository

	// Attempt ClickHouse connection, fallback to stateful memory repo if offline
	chAdapter, err := clickhouse.NewAdapter(cfg.ClickHouse, log)
	if err == nil {
		defer chAdapter.Close()
		metricsRepo = chAdapter
		ineffRepo = clickhouse.NewInefficiencyAdapter(chAdapter.Conn(), log)
		recRepo = clickhouse.NewRecommendationAdapter(chAdapter.Conn(), log)
		alertRepo = clickhouse.NewAlertAdapter(chAdapter.Conn(), log)

		healthRegistry.Register("clickhouse", func(ctx context.Context) error {
			return chAdapter.Ping(ctx)
		})
		log.Info().Msg("using ClickHouse OLAP repository")
	} else {
		log.Warn().Err(err).Msg("ClickHouse unavailable — initializing stateful in-memory repository")
		memStore := memoryadapter.NewRepositoryStore(log)
		metricsRepo = memStore
		ineffRepo = memStore.InefficiencyRepo()
		recRepo = memStore.RecommendationRepo()
		alertRepo = memStore.AlertRepo()
		authRepo = memStore.AuthRepo()
		apiKeyRepo = memStore.APIKeyRepo()
		auditRepo = memStore.AuditLogRepo()
		usageRepo = memStore.UsageRepo()
	}

	var redisClient *redis.Client
	redisAdapter, err := redisadapter.NewAdapter(cfg.Redis, log)
	if err == nil {
		defer redisAdapter.Close()
		redisClient = redisAdapter.Client()
		healthRegistry.Register("redis", func(ctx context.Context) error {
			return redisAdapter.Ping(ctx)
		})
	} else {
		log.Warn().Err(err).Msg("Redis unavailable — continuing without Redis rate-limiting cache")
	}

	// Initialize services
	ingestionSvc := ingestion.NewService(
		metricsRepo,
		redisAdapter,
		log,
		cfg.ClickHouse.BatchSize,
		cfg.ClickHouse.FlushInterval,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ingestionSvc.StartPeriodicFlush(ctx)

	// Setup Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "obsygnal-pulse-secret-dev-key-32b"
	}

	authService := authsvc.NewService(
		jwtSecret,
		cfg.Auth.JWTExpiry,
		log,
		authsvc.WithAuthRepo(authRepo),
		authsvc.WithAPIKeyRepo(apiKeyRepo),
		authsvc.WithAuditRepo(auditRepo),
		authsvc.WithUsageRepo(usageRepo),
	)

	router := gin.New()

	routes.Register(router, routes.Deps{
		Log:          log,
		Registry:     healthRegistry,
		MetricsRepo:  metricsRepo,
		IneffRepo:    ineffRepo,
		RecRepo:      recRepo,
		AlertRepo:    alertRepo,
		UsageRepo:    usageRepo,
		IngestionSvc: ingestionSvc,
		RedisClient:  redisClient,
		AuthService:  authService,
		CORSOrigins:  cfg.Server.CORSAllowedOrigins,
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

	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info().Str("signal", sig.String()).Msg("shutting down API server")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}

	ingestionSvc.Flush(context.Background())
	log.Info().Msg("server stopped cleanly")
}
