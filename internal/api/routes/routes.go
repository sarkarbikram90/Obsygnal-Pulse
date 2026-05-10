// Package routes centralizes the HTTP routing topology for the api-server.
//
// It strictly maps URL paths (e.g., /api/v1/health) to their respective handlers
// and wires up critical middleware (Rate Limiting, JWT Authentication, CORS).
// This separation ensures the HTTP server configuration is cleanly decoupled
// from the concrete endpoint logic.
package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/handlers"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/middleware"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

// Deps holds the dependencies needed for route registration.
type Deps struct {
	Log          *logger.Logger
	Registry     *health.Registry
	MetricsRepo  ports.MetricsRepository
	IngestionSvc *ingestion.Service
	RedisClient  *redis.Client
	JWTSecret    string
}

// Register sets up all API routes on the provided Gin engine.
func Register(r *gin.Engine, deps Deps) {
	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())

	// Apply rate limiting if redis is provided
	if deps.RedisClient != nil {
		r.Use(middleware.RateLimiter(deps.RedisClient, deps.Log, 100, 1*time.Minute))
	}

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Health & status (Public)
		v1.GET("/health", handlers.HealthCheck(deps.Registry))
		v1.GET("/ready", handlers.ReadinessCheck(deps.Registry))

		// Ingestion (Public for now, could be protected via API key later)
		v1.POST("/metrics/ingest", handlers.IngestMetrics(deps.IngestionSvc))
		v1.GET("/ingestion/stats", handlers.IngestionStats(deps.IngestionSvc))

		// Protected endpoints
		protected := v1.Group("")
		if deps.JWTSecret != "" {
			protected.Use(middleware.AuthRequired(deps.Log, deps.JWTSecret))
		}

		{
			// Cluster & Metrics
			protected.GET("/cluster/overview", handlers.ClusterOverview(deps.MetricsRepo))
			protected.GET("/gpus/metrics", handlers.GPUMetrics(deps.MetricsRepo))

			// Inefficiencies
			protected.GET("/inefficiencies", handlers.ListInefficiencies())

			// Recommendations
			protected.GET("/recommendations", handlers.ListRecommendations())
			protected.POST("/recommendations/:id/apply", handlers.ApplyRecommendation())
			protected.POST("/recommendations/:id/dismiss", handlers.DismissRecommendation())

			// Alerts
			protected.GET("/alerts/recent", handlers.RecentAlerts())
		}
	}
}
