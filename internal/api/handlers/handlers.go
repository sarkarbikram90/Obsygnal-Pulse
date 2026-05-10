// Package handlers provides the concrete HTTP endpoint handlers for the api-server.
// Handlers act strictly as the Primary Adapter in the Hexagonal Architecture:
// they parse JSON requests into Domain Transfer Objects (DTOs), invoke the
// corresponding internal business logic Services, and format the resulting
// Domain models back into JSON responses. No business rules or direct
// database queries are permitted in this package.
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

// HealthCheck returns a handler for health check endpoints.
func HealthCheck(registry *health.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		report := registry.Check(c.Request.Context())
		status := http.StatusOK
		if report.Status == health.StatusUnhealthy {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, report)
	}
}

// ReadinessCheck returns a handler for readiness probes.
func ReadinessCheck(registry *health.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		report := registry.Check(c.Request.Context())
		if report.Status == health.StatusUnhealthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ready": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	}
}

// ClusterOverview returns a handler for cluster overview data.
func ClusterOverview(repo ports.MetricsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.DefaultQuery("cluster_id", "sim-cluster-01")
		overview, err := repo.GetClusterOverview(c.Request.Context(), clusterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, overview)
	}
}

// GPUMetrics returns a handler for querying GPU metrics.
func GPUMetrics(repo ports.MetricsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := ports.MetricsQuery{
			ClusterID: c.Query("cluster_id"),
			NodeID:    c.Query("node_id"),
			GPUID:     c.Query("gpu_id"),
			JobID:     c.Query("job_id"),
			Limit:     100,
		}

		// Parse time range
		if from := c.Query("from"); from != "" {
			if t, err := time.Parse(time.RFC3339, from); err == nil {
				query.From = t
			}
		}
		if to := c.Query("to"); to != "" {
			if t, err := time.Parse(time.RFC3339, to); err == nil {
				query.To = t
			}
		}

		// Default to last hour
		if query.From.IsZero() {
			query.From = time.Now().Add(-1 * time.Hour)
		}
		if query.To.IsZero() {
			query.To = time.Now()
		}

		results, err := repo.QueryGPUMetrics(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  results,
			"count": len(results),
			"query": query,
		})
	}
}

// IngestMetrics returns a handler for receiving metric batches via API.
func IngestMetrics(svc *ingestion.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var batch metrics.MetricBatch
		if err := c.ShouldBindJSON(&batch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}

		batch.ReceivedAt = time.Now()
		if batch.Source == "" {
			batch.Source = "api"
		}

		if err := svc.ProcessBatch(c.Request.Context(), &batch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"accepted":    batch.Size(),
			"received_at": batch.ReceivedAt,
		})
	}
}

// IngestionStats returns a handler for ingestion pipeline statistics.
func IngestionStats(svc *ingestion.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, svc.Stats())
	}
}
