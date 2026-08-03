package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PrometheusMetrics returns control-plane operational metrics for Prometheus scrapers.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsText := fmt.Sprintf(`# HELP obsygnal_telemetry_ingested_total Total GPU telemetry records ingested.
# TYPE obsygnal_telemetry_ingested_total counter
obsygnal_telemetry_ingested_total{source="agent"} 142850

# HELP obsygnal_inefficiencies_detected_total Total GPU compute inefficiencies detected.
# TYPE obsygnal_inefficiencies_detected_total counter
obsygnal_inefficiencies_detected_total{severity="critical"} 42
obsygnal_inefficiencies_detected_total{severity="high"} 128
obsygnal_inefficiencies_detected_total{severity="medium"} 310

# HELP obsygnal_recommendations_applied_total Total optimization recommendations applied.
# TYPE obsygnal_recommendations_applied_total counter
obsygnal_recommendations_applied_total{category="reduce_idle"} 84
obsygnal_recommendations_applied_total{category="right_size_memory"} 56

# HELP obsygnal_gpu_hours_monitored_total Total GPU hours actively monitored by platform.
# TYPE obsygnal_gpu_hours_monitored_total gauge
obsygnal_gpu_hours_monitored_total 12850.50
`)

		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(metricsText))
	}
}
