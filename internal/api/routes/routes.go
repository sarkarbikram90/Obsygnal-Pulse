// Package routes registers all API routes on the Gin engine.
//
// Route registration is separated from handlers for clarity:
// - routes.go defines URL patterns and maps them to handlers
// - handlers.go implements the request/response logic
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/obsygnal/obsygnal-pulse/internal/api/handlers"
	"github.com/obsygnal/obsygnal-pulse/internal/api/middleware"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/health"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
	"github.com/obsygnal/obsygnal-pulse/internal/ports"
	"github.com/obsygnal/obsygnal-pulse/internal/services/ingestion"
)

// Deps holds the dependencies needed for route registration.
type Deps struct {
	Log          *logger.Logger
	Registry     *health.Registry
	MetricsRepo  ports.MetricsRepository
	IngestionSvc *ingestion.Service
}

// Register sets up all API routes on the provided Gin engine.
func Register(r *gin.Engine, deps Deps) {
	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Health & status
		v1.GET("/health", handlers.HealthCheck(deps.Registry))
		v1.GET("/ready", handlers.ReadinessCheck(deps.Registry))

		// Cluster overview
		v1.GET("/cluster/overview", handlers.ClusterOverview(deps.MetricsRepo))

		// GPU metrics
		v1.GET("/gpus/metrics", handlers.GPUMetrics(deps.MetricsRepo))

		// Ingestion
		v1.POST("/metrics/ingest", handlers.IngestMetrics(deps.IngestionSvc))
		v1.GET("/ingestion/stats", handlers.IngestionStats(deps.IngestionSvc))
	}
}
