// Package routes centralizes the HTTP routing topology for the api-server.
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
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

// Deps holds the dependencies needed for route registration.
type Deps struct {
	Log          *logger.Logger
	Registry     *health.Registry
	MetricsRepo  ports.MetricsRepository
	IneffRepo    ports.InefficiencyRepository
	RecRepo      ports.RecommendationRepository
	AlertRepo    ports.AlertRepository
	UsageRepo    ports.UsageRepository
	IngestionSvc *ingestion.Service
	RedisClient  *redis.Client
	AuthService  *authsvc.Service
	CORSOrigins  string
}

// Register sets up all API routes on the provided Gin engine.
func Register(r *gin.Engine, deps Deps) {
	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.CORS(deps.CORSOrigins))
	r.Use(middleware.SecurityHeaders())

	// Apply rate limiting if redis is provided
	if deps.RedisClient != nil {
		r.Use(middleware.RateLimiter(deps.RedisClient, deps.Log, 200, 1*time.Minute))
	}

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Health & status (Public)
		v1.GET("/health", handlers.HealthCheck(deps.Registry))
		v1.GET("/ready", handlers.ReadinessCheck(deps.Registry))
		r.GET("/metrics", handlers.PrometheusMetrics())

		// Authentication (Public)
		v1.POST("/auth/register", handlers.RegisterUser(deps.AuthService))
		v1.POST("/auth/login", handlers.LoginUser(deps.AuthService))

		// Ingestion (Public for agent metric ingestion)
		v1.POST("/metrics/ingest", handlers.IngestMetrics(deps.IngestionSvc))
		v1.GET("/ingestion/stats", handlers.IngestionStats(deps.IngestionSvc))

		// Protected endpoints (Supports JWT or fallback for dev)
		protected := v1.Group("")
		if deps.AuthService != nil {
			protected.Use(middleware.AuthRequired(deps.Log, deps.AuthService))
		}

		{
			// Auth & Organization
			protected.GET("/auth/me", handlers.GetCurrentUser())

			// Cluster & Metrics
			protected.GET("/cluster/overview", handlers.ClusterOverview(deps.MetricsRepo))
			protected.GET("/gpus/metrics", handlers.GPUMetrics(deps.MetricsRepo))

			// Inefficiencies & Bottlenecks
			protected.GET("/inefficiencies", handlers.ListInefficiencies(deps.IneffRepo))

			// Recommendations & Auto-Remediation
			protected.GET("/recommendations", handlers.ListRecommendations(deps.RecRepo))
			protected.POST("/recommendations/:id/apply", handlers.ApplyRecommendation(deps.RecRepo))
			protected.POST("/recommendations/:id/dismiss", handlers.DismissRecommendation(deps.RecRepo))

			// Alerts
			protected.GET("/alerts/recent", handlers.RecentAlerts(deps.AlertRepo))

			// Training Jobs & Nodes
			protected.GET("/jobs", handlers.ListJobs())
			protected.GET("/nodes", handlers.ListNodes())

			// API Keys
			protected.POST("/api-keys", handlers.CreateAPIKey(deps.AuthService))
			protected.GET("/api-keys", handlers.ListAPIKeys(deps.AuthService))
			protected.DELETE("/api-keys/:id", handlers.RevokeAPIKey(deps.AuthService))

			// SaaS Billing & Metering
			protected.GET("/billing/usage", handlers.GetUsageSummary(deps.UsageRepo))
			protected.POST("/billing/upgrade", handlers.UpgradeTier())

			// Compliance Audit Logs
			protected.GET("/audit-logs", handlers.ListAuditLogs(deps.AuthService))
		}
	}
}
