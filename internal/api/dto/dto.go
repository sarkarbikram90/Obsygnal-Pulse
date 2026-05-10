// Package dto defines Data Transfer Objects for the API layer.
//
// DTOs handle the conversion between HTTP request/response formats and
// internal domain models. This keeps transport concerns out of the domain.
package dto

import (
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/detection"
	"github.com/obsygnal/obsygnal-pulse/internal/domain/recommendation"
)

// MetricsIngestRequest is the request body for POST /api/v1/metrics/ingest.
type MetricsIngestRequest struct {
	GPUMetrics []GPUMetricDTO  `json:"gpu_metrics"`
	NodeMetrics []NodeMetricDTO `json:"node_metrics,omitempty"`
	Source      string          `json:"source"`
}

// GPUMetricDTO is the wire format for a GPU metric data point.
type GPUMetricDTO struct {
	ClusterID         string            `json:"cluster_id" binding:"required"`
	NodeID            string            `json:"node_id" binding:"required"`
	GPUID             string            `json:"gpu_id" binding:"required"`
	GPUModel          string            `json:"gpu_model,omitempty"`
	GPUIndex          int               `json:"gpu_index,omitempty"`
	Timestamp         time.Time         `json:"timestamp" binding:"required"`
	Utilization       float64           `json:"utilization"`
	MemoryUtilization float64           `json:"memory_utilization"`
	MemoryUsed        uint64            `json:"memory_used,omitempty"`
	MemoryTotal       uint64            `json:"memory_total,omitempty"`
	Temperature       float64           `json:"temperature,omitempty"`
	PowerUsage        float64           `json:"power_usage,omitempty"`
	RunningProcesses  int               `json:"running_processes,omitempty"`
	JobID             string            `json:"job_id,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// NodeMetricDTO is the wire format for a node metric data point.
type NodeMetricDTO struct {
	ClusterID   string  `json:"cluster_id" binding:"required"`
	NodeID      string  `json:"node_id" binding:"required"`
	Timestamp   time.Time `json:"timestamp" binding:"required"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	GPUCount    int     `json:"gpu_count"`
}

// InefficiencyResponse is the response format for inefficiency endpoints.
type InefficiencyResponse struct {
	Data       []detection.Inefficiency `json:"data"`
	TotalCount int                      `json:"total_count"`
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
}

// RecommendationResponse is the response format for recommendation endpoints.
type RecommendationResponse struct {
	Data       []recommendation.Recommendation `json:"data"`
	TotalCount int                              `json:"total_count"`
	Limit      int                              `json:"limit"`
	Offset     int                              `json:"offset"`
}

// DismissRequest is the request body for POST /api/v1/recommendations/:id/dismiss.
type DismissRequest struct {
	DismissedBy string `json:"dismissed_by" binding:"required"`
	Reason      string `json:"reason" binding:"required"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error specifics.
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// PaginationParams extracts pagination from query parameters.
type PaginationParams struct {
	Limit  int `form:"limit,default=50" binding:"min=1,max=1000"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}
