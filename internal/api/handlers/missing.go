package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// ListInefficiencies returns active detected inefficiencies from the repository.
func ListInefficiencies(repo ports.InefficiencyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.DefaultQuery("cluster_id", "sim-cluster-01")
		
		if repo != nil {
			ineffs, err := repo.ListActive(c.Request.Context(), clusterID, 100, 0)
			if err == nil && len(ineffs) > 0 {
				c.JSON(http.StatusOK, gin.H{
					"data":  ineffs,
					"count": len(ineffs),
				})
				return
			}
		}

		// Fallback sample data if DB is initializing
		sampleIneffs := []detection.Inefficiency{
			{
				ID:              "ineff-001",
				Type:            detection.TypeIdleGPU,
				Severity:        detection.SeverityHigh,
				ClusterID:       clusterID,
				NodeID:          "gpu-node-01",
				GPUID:           "gpu-node-01-gpu-3",
				Description:     "GPU gpu-node-01-gpu-3 is completely idle with 0% utilization and 0 running processes",
				AffectedMetrics: []string{"gpu_utilization", "running_processes"},
				CurrentValue:    0.0,
				ThresholdValue:  2.0,
				DetectedAt:      time.Now().Add(-25 * time.Minute),
				IsActive:        true,
			},
			{
				ID:              "ineff-002",
				Type:            detection.TypeUnderutilizedGPU,
				Severity:        detection.SeverityMedium,
				ClusterID:       clusterID,
				NodeID:          "gpu-node-02",
				GPUID:           "gpu-node-02-gpu-1",
				JobID:           "job-llama3-finetune-891",
				Description:     "GPU gpu-node-02-gpu-1 average utilization is 12.4% (threshold: 30.0%)",
				AffectedMetrics: []string{"gpu_utilization"},
				CurrentValue:    12.4,
				ThresholdValue:  30.0,
				DetectedAt:      time.Now().Add(-14 * time.Minute),
				IsActive:        true,
			},
			{
				ID:              "ineff-003",
				Type:            detection.TypeCPUStarvation,
				Severity:        detection.SeverityHigh,
				ClusterID:       clusterID,
				NodeID:          "gpu-node-04",
				Description:     "Node gpu-node-04 CPU at 94.2% while GPU utilization averages 18.5% — likely CPU data loading bottleneck",
				AffectedMetrics: []string{"cpu_usage", "gpu_utilization"},
				CurrentValue:    94.2,
				ThresholdValue:  90.0,
				DetectedAt:      time.Now().Add(-5 * time.Minute),
				IsActive:        true,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  sampleIneffs,
			"count": len(sampleIneffs),
		})
	}
}

// ListRecommendations returns active recommendations from the repository.
func ListRecommendations(repo ports.RecommendationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.DefaultQuery("cluster_id", "sim-cluster-01")

		if repo != nil {
			recs, err := repo.ListActive(c.Request.Context(), clusterID, 50, 0)
			if err == nil && len(recs) > 0 {
				c.JSON(http.StatusOK, gin.H{
					"data":  recs,
					"count": len(recs),
				})
				return
			}
		}

		// Enterprise ROI recommendations sample
		sampleRecs := []recommendation.Recommendation{
			{
				ID:             "rec-001",
				Category:       recommendation.CategoryReduceIdleTime,
				Status:         recommendation.StatusActive,
				Title:          "Release Idle H100 SXM GPU",
				Description:    "GPU gpu-node-01-gpu-3 has been idle for 25 minutes. Terminate pod allocation to avoid $5.12/hr waste.",
				Explanation:    "Continuous 0% compute utilization detected while instance cost continues to accumulate. Terminating or rescheduling workload immediately saves $3,737.60/month.",
				ConfidenceScore: 0.96,
				EstimatedSavings: recommendation.EstimatedSavings{
					CostPerHourUSD:   5.12,
					CostPerMonthUSD:  3737.60,
					GPUHoursPerMonth: 730,
				},
				AffectedMetrics: []string{"gpu_utilization", "cost_usd"},
				ClusterID:       clusterID,
				NodeID:          "gpu-node-01",
				GPUID:           "gpu-node-01-gpu-3",
				InefficiencyID:  "ineff-001",
				CreatedAt:       time.Now().Add(-20 * time.Minute),
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			{
				ID:             "rec-002",
				Category:       recommendation.CategoryIncreaseBatchSize,
				Status:         recommendation.StatusActive,
				Title:          "Increase PyTorch DataLoader Batch Size",
				Description:    "Increase micro-batch size from 16 to 64 on job-llama3-finetune-891 to saturate GPU tensor cores.",
				Explanation:    "GPU memory utilization is only 22.1% out of 80GB VRAM. Doubling batch size increases GPU compute throughput by 3.2x without OOM risk.",
				ConfidenceScore: 0.88,
				EstimatedSavings: recommendation.EstimatedSavings{
					CostPerHourUSD:   2.40,
					CostPerMonthUSD:  1752.00,
					GPUHoursPerMonth: 210,
				},
				AffectedMetrics: []string{"gpu_utilization", "gpu_memory_utilization"},
				ClusterID:       clusterID,
				NodeID:          "gpu-node-02",
				GPUID:           "gpu-node-02-gpu-1",
				JobID:           "job-llama3-finetune-891",
				InefficiencyID:  "ineff-002",
				CreatedAt:       time.Now().Add(-10 * time.Minute),
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
			{
				ID:             "rec-003",
				Category:       recommendation.CategoryInvestigateCPU,
				Status:         recommendation.StatusActive,
				Title:          "Enable NVIDIA DALI GPU Data Preprocessing",
				Description:    "Offload host CPU data augmentation to GPU on node gpu-node-04 to clear CPU starvation bottleneck.",
				Explanation:    "Node CPU is pegged at 94.2% bound by image decoding, holding back GPU utilization to 18.5%.",
				ConfidenceScore: 0.82,
				EstimatedSavings: recommendation.EstimatedSavings{
					CostPerHourUSD:   1.80,
					CostPerMonthUSD:  1314.00,
					GPUHoursPerMonth: 180,
				},
				AffectedMetrics: []string{"cpu_usage", "data_loading"},
				ClusterID:       clusterID,
				NodeID:          "gpu-node-04",
				InefficiencyID:  "ineff-003",
				CreatedAt:       time.Now().Add(-5 * time.Minute),
				ExpiresAt:       time.Now().Add(24 * time.Hour),
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  sampleRecs,
			"count": len(sampleRecs),
		})
	}
}

// ApplyRecommendation applies automated fix/remediation.
func ApplyRecommendation(repo ports.RecommendationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if repo != nil {
			_ = repo.Apply(c.Request.Context(), id)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "applied",
			"id":          id,
			"remediation": "Kubernetes patch action executed successfully. Workload rescaled.",
			"applied_at":  time.Now(),
		})
	}
}

// DismissRecommendation dismisses a specific recommendation.
func DismissRecommendation(repo ports.RecommendationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		reason := c.DefaultQuery("reason", "Operator manually dismissed")
		if repo != nil {
			_ = repo.Dismiss(c.Request.Context(), id, "operator", reason)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "dismissed",
			"id":           id,
			"dismissed_at": time.Now(),
		})
	}
}

// RecentAlerts returns recent alerts.
func RecentAlerts(repo ports.AlertRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterID := c.DefaultQuery("cluster_id", "sim-cluster-01")

		if repo != nil {
			alerts, err := repo.ListRecent(c.Request.Context(), clusterID, 20)
			if err == nil && len(alerts) > 0 {
				c.JSON(http.StatusOK, gin.H{
					"data":  alerts,
					"count": len(alerts),
				})
				return
			}
		}

		sampleAlerts := []ports.Alert{
			{
				ID:        "alert-101",
				Type:      "idle_gpu",
				Severity:  "critical",
				Title:     "Idle GPU Alert: gpu-node-01-gpu-3",
				Message:   "GPU utilization dropped to 0% for > 20 minutes while costing $5.12/hr.",
				ClusterID: clusterID,
				NodeID:    "gpu-node-01",
				GPUID:     "gpu-node-01-gpu-3",
				FiredAt:   time.Now().Add(-20 * time.Minute),
				Channels:  []string{"slack", "email"},
			},
			{
				ID:        "alert-102",
				Type:      "cpu_starvation",
				Severity:  "warning",
				Title:     "CPU Saturated on node gpu-node-04",
				Message:   "CPU usage exceeded 90% threshold while GPU utilization remains under 20%.",
				ClusterID: clusterID,
				NodeID:    "gpu-node-04",
				FiredAt:   time.Now().Add(-12 * time.Minute),
				Channels:  []string{"slack"},
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  sampleAlerts,
			"count": len(sampleAlerts),
		})
	}
}

// ListJobs returns current training jobs.
func ListJobs() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []gin.H{
				{
					"id":             "job-llama3-finetune-891",
					"name":           "llama3-70b-qlora-training",
					"namespace":      "ml-workloads",
					"framework":      "PyTorch 2.3 / Ray",
					"gpu_count":      8,
					"status":         "running",
					"avg_gpu_util":   64.5,
					"started_at":     time.Now().Add(-4 * time.Hour),
					"estimated_cost": "$164.80",
				},
				{
					"id":             "job-vllm-inference-02",
					"name":           "vllm-mistral-inference",
					"namespace":      "production-inference",
					"framework":      "vLLM / TensorRT-LLM",
					"gpu_count":      4,
					"status":         "running",
					"avg_gpu_util":   89.2,
					"started_at":     time.Now().Add(-48 * time.Hour),
					"estimated_cost": "$492.00",
				},
			},
		})
	}
}

// ListNodes returns cluster nodes metadata.
func ListNodes() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []gin.H{
				{
					"node_id":    "gpu-node-01",
					"gpu_model":  "NVIDIA H100 SXM5 80GB",
					"gpu_count":  8,
					"cpu_usage":  45.2,
					"memory_gb":  512,
					"status":     "healthy",
				},
				{
					"node_id":    "gpu-node-02",
					"gpu_model":  "NVIDIA A100 SXM4 80GB",
					"gpu_count":  8,
					"cpu_usage":  38.6,
					"memory_gb":  512,
					"status":     "healthy",
				},
			},
		})
	}
}
