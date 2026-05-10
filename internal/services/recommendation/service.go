// Package recommendation implements the automated optimization recommendation engine.
// It translates mathematical inefficiencies (identified by the detection engine) into
// human-actionable infrastructure modifications (e.g., "Downscale to T4 GPU"), attaching
// calculated confidence scores and estimated monetary savings.
// Architecture:
//
//	The recommendation engine transforms detected inefficiencies into actionable
//	optimization suggestions. Each inefficiency type maps to one or more
//	recommendation generators. Each generator produces a Recommendation with:
//	- A confidence score (0.0-1.0) based on data quality and pattern strength
//	- Estimated cost savings (hourly and monthly)
//	- Affected metrics for drill-down
//	- Human-readable explanation
//
// Design decisions:
//   - Recommendations are generated reactively from detection events (not polled)
//   - Confidence scores are rule-based for MVP, ML-based in future
//   - Cost estimates use configurable GPU pricing (default: A100 on-demand)
//   - Recommendations expire after a configurable TTL to prevent stale advice
//   - Deduplication prevents duplicate recommendations for the same inefficiency
package recommendation

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/recommendation"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// GPUPricing holds cost data for GPU types.
type GPUPricing struct {
	Model          string
	CostPerHourUSD float64
	MemoryGB       float64
	TDPWatts       float64
}

// DefaultPricing returns pricing for common GPU types.
func DefaultPricing() map[string]GPUPricing {
	return map[string]GPUPricing{
		"NVIDIA A100-SXM4-80GB": {Model: "A100-80GB", CostPerHourUSD: 3.67, MemoryGB: 80, TDPWatts: 400},
		"NVIDIA A100-SXM4-40GB": {Model: "A100-40GB", CostPerHourUSD: 2.21, MemoryGB: 40, TDPWatts: 400},
		"NVIDIA H100-SXM5-80GB": {Model: "H100-80GB", CostPerHourUSD: 5.12, MemoryGB: 80, TDPWatts: 700},
		"NVIDIA V100-SXM2-32GB": {Model: "V100-32GB", CostPerHourUSD: 1.46, MemoryGB: 32, TDPWatts: 300},
		"default":               {Model: "Generic", CostPerHourUSD: 2.50, MemoryGB: 40, TDPWatts: 350},
	}
}

// Service generates optimization recommendations from detected inefficiencies.
type Service struct {
	recRepo ports.RecommendationRepository
	cache   ports.CacheStore
	log     *logger.Logger
	pricing map[string]GPUPricing
	recTTL  time.Duration

	totalGenerated int64
	totalDismissed int64
}

// NewService creates a new recommendation service.
func NewService(
	recRepo ports.RecommendationRepository,
	cache ports.CacheStore,
	log *logger.Logger,
) *Service {
	return &Service{
		recRepo: recRepo,
		cache:   cache,
		log:     log.WithComponent("recommendation"),
		pricing: DefaultPricing(),
		recTTL:  24 * time.Hour,
	}
}

// GenerateFromInefficiency creates recommendations based on a detected inefficiency.
// This is the main entry point — called by the detection engine's OnDetection callback.
func (s *Service) GenerateFromInefficiency(ctx context.Context, ineff detection.Inefficiency) ([]recommendation.Recommendation, error) {
	generators := s.getGenerators(ineff.Type)
	if len(generators) == 0 {
		return nil, nil
	}

	var recs []recommendation.Recommendation
	for _, gen := range generators {
		rec := gen(ineff)
		if rec != nil {
			rec.ID = uuid.New().String()
			rec.Status = recommendation.StatusActive
			rec.CreatedAt = time.Now()
			rec.ExpiresAt = time.Now().Add(s.recTTL)
			rec.InefficiencyID = ineff.ID

			recs = append(recs, *rec)
		}
	}

	if len(recs) > 0 {
		if err := s.recRepo.InsertBatch(ctx, recs); err != nil {
			s.log.Error().Err(err).Int("count", len(recs)).Msg("failed to persist recommendations")
			return recs, err
		}
		s.totalGenerated += int64(len(recs))

		s.log.Info().
			Int("count", len(recs)).
			Str("inefficiency_type", string(ineff.Type)).
			Str("gpu_id", ineff.GPUID).
			Msg("generated recommendations")
	}

	return recs, nil
}

// Dismiss marks a recommendation as dismissed with a reason.
func (s *Service) Dismiss(ctx context.Context, id, dismissedBy, reason string) error {
	if err := s.recRepo.Dismiss(ctx, id, dismissedBy, reason); err != nil {
		return fmt.Errorf("failed to dismiss recommendation: %w", err)
	}
	s.totalDismissed++
	return nil
}

// ListActive returns active recommendations for a cluster.
func (s *Service) ListActive(ctx context.Context, clusterID string, limit, offset int) ([]recommendation.Recommendation, error) {
	return s.recRepo.ListActive(ctx, clusterID, limit, offset)
}

// Stats returns recommendation engine statistics.
func (s *Service) Stats() map[string]int64 {
	return map[string]int64{
		"total_generated": s.totalGenerated,
		"total_dismissed": s.totalDismissed,
	}
}

// getGenerators returns the recommendation generator functions for a given inefficiency type.
func (s *Service) getGenerators(itype detection.InefficiencyType) []func(detection.Inefficiency) *recommendation.Recommendation {
	switch itype {
	case detection.TypeUnderutilizedGPU:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genReduceAllocation,
			s.genIncreaseBatchSize,
		}
	case detection.TypeIdleGPU:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genReduceIdleTime,
		}
	case detection.TypeLowMemoryUsage:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genRightSizeMemory,
			s.genIncreaseBatchSize,
		}
	case detection.TypeCPUStarvation:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genInvestigateCPU,
		}
	case detection.TypeStalledTraining:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genInvestigateStalledTraining,
		}
	case detection.TypeAbnormalSpike:
		return []func(detection.Inefficiency) *recommendation.Recommendation{
			s.genInvestigateThermal,
		}
	default:
		return nil
	}
}

// ============================================================
// RECOMMENDATION GENERATORS
// Each generator creates a specific type of recommendation from
// an inefficiency, including confidence score and savings estimate.
// ============================================================

func (s *Service) genReduceAllocation(ineff detection.Inefficiency) *recommendation.Recommendation {
	pricing := s.getGPUPricing(ineff.Labels)
	wastePercent := (ineff.ThresholdValue - ineff.CurrentValue) / ineff.ThresholdValue
	confidence := s.calculateConfidence(ineff.CurrentValue, ineff.ThresholdValue, 0.3)

	savings := recommendation.EstimatedSavings{
		CostPerHourUSD:   pricing.CostPerHourUSD * wastePercent,
		CostPerMonthUSD:  pricing.CostPerHourUSD * wastePercent * 730, // ~730 hours/month
		GPUHoursPerMonth: 730 * wastePercent,
	}

	return &recommendation.Recommendation{
		Category:         recommendation.CategoryReduceAllocation,
		Title:            fmt.Sprintf("Reduce GPU allocation for %s", ineff.GPUID),
		Description:      fmt.Sprintf("GPU %s is averaging %.1f%% utilization. Consider reducing the GPU count or switching to a smaller GPU type.", ineff.GPUID, ineff.CurrentValue),
		Explanation:      fmt.Sprintf("Over the evaluation window, GPU compute utilization averaged %.1f%%, which is %.1f percentage points below the %.1f%% threshold. This suggests the workload does not fully utilize the allocated GPU resources. Reducing allocation would save approximately $%.2f/month.", ineff.CurrentValue, ineff.ThresholdValue-ineff.CurrentValue, ineff.ThresholdValue, savings.CostPerMonthUSD),
		ConfidenceScore:  confidence,
		EstimatedSavings: savings,
		AffectedMetrics:  []string{"gpu_utilization"},
		ClusterID:        ineff.ClusterID,
		NodeID:           ineff.NodeID,
		GPUID:            ineff.GPUID,
		JobID:            ineff.JobID,
	}
}

func (s *Service) genIncreaseBatchSize(ineff detection.Inefficiency) *recommendation.Recommendation {
	confidence := s.calculateConfidence(ineff.CurrentValue, ineff.ThresholdValue, 0.2)

	return &recommendation.Recommendation{
		Category:        recommendation.CategoryIncreaseBatchSize,
		Title:           fmt.Sprintf("Increase batch size for workload on %s", ineff.GPUID),
		Description:     fmt.Sprintf("GPU %s has low utilization (%.1f%%). Increasing the training batch size could improve GPU throughput.", ineff.GPUID, ineff.CurrentValue),
		Explanation:     "Low GPU utilization often indicates the batch size is too small to saturate the GPU compute pipeline. Increasing batch size typically improves training throughput linearly until GPU memory becomes the bottleneck.",
		ConfidenceScore: confidence,
		EstimatedSavings: recommendation.EstimatedSavings{
			GPUHoursPerMonth: 50, // Conservative estimate
		},
		AffectedMetrics: []string{"gpu_utilization", "training_throughput"},
		ClusterID:       ineff.ClusterID,
		NodeID:          ineff.NodeID,
		GPUID:           ineff.GPUID,
		JobID:           ineff.JobID,
	}
}

func (s *Service) genReduceIdleTime(ineff detection.Inefficiency) *recommendation.Recommendation {
	pricing := s.getGPUPricing(ineff.Labels)
	savings := recommendation.EstimatedSavings{
		CostPerHourUSD:   pricing.CostPerHourUSD,
		CostPerMonthUSD:  pricing.CostPerHourUSD * 730,
		GPUHoursPerMonth: 730,
	}

	return &recommendation.Recommendation{
		Category:         recommendation.CategoryReduceIdleTime,
		Title:            fmt.Sprintf("Idle GPU detected: %s", ineff.GPUID),
		Description:      fmt.Sprintf("GPU %s is completely idle with no running processes. Release it or assign a workload.", ineff.GPUID),
		Explanation:      "This GPU is consuming power and cloud costs without performing any computation. If no workload is planned, releasing this GPU allocation will eliminate waste immediately.",
		ConfidenceScore:  0.95, // High confidence for idle GPUs
		EstimatedSavings: savings,
		AffectedMetrics:  []string{"gpu_utilization", "running_processes"},
		ClusterID:        ineff.ClusterID,
		NodeID:           ineff.NodeID,
		GPUID:            ineff.GPUID,
	}
}

func (s *Service) genRightSizeMemory(ineff detection.Inefficiency) *recommendation.Recommendation {
	pricing := s.getGPUPricing(ineff.Labels)
	memWastePercent := (100 - ineff.CurrentValue) / 100

	return &recommendation.Recommendation{
		Category:        recommendation.CategoryRightSizeMemory,
		Title:           fmt.Sprintf("Right-size GPU memory for %s", ineff.GPUID),
		Description:     fmt.Sprintf("GPU %s is only using %.1f%% of available memory. A GPU with less memory would suffice.", ineff.GPUID, ineff.CurrentValue),
		Explanation:     fmt.Sprintf("Memory utilization at %.1f%% means %.0f%% of allocated GPU memory is wasted. Moving to a smaller GPU (e.g., 40GB instead of 80GB) could reduce costs while maintaining performance.", ineff.CurrentValue, memWastePercent*100),
		ConfidenceScore: s.calculateConfidence(ineff.CurrentValue, ineff.ThresholdValue, 0.25),
		EstimatedSavings: recommendation.EstimatedSavings{
			CostPerHourUSD:  pricing.CostPerHourUSD * 0.4, // Approximate savings from downsizing
			CostPerMonthUSD: pricing.CostPerHourUSD * 0.4 * 730,
			MemoryGB:        pricing.MemoryGB * memWastePercent,
		},
		AffectedMetrics: []string{"gpu_memory_utilization"},
		ClusterID:       ineff.ClusterID,
		NodeID:          ineff.NodeID,
		GPUID:           ineff.GPUID,
		JobID:           ineff.JobID,
	}
}

func (s *Service) genInvestigateCPU(ineff detection.Inefficiency) *recommendation.Recommendation {
	return &recommendation.Recommendation{
		Category:        recommendation.CategoryInvestigateCPU,
		Title:           fmt.Sprintf("Investigate CPU bottleneck on %s", ineff.NodeID),
		Description:     fmt.Sprintf("Node %s has CPU usage at %.1f%% while GPUs are underutilized. Data preprocessing or loading may be the bottleneck.", ineff.NodeID, ineff.CurrentValue),
		Explanation:     "When CPU is saturated but GPUs are idle or underutilized, the CPU is likely a bottleneck — commonly from data loading, preprocessing, or augmentation. Solutions include: increasing dataloader workers, using NVIDIA DALI for GPU-accelerated preprocessing, pre-processing data offline, or adding CPU resources.",
		ConfidenceScore: 0.80,
		EstimatedSavings: recommendation.EstimatedSavings{
			GPUHoursPerMonth: 200, // Potential recovery from idle GPU time
		},
		AffectedMetrics: []string{"cpu_usage", "gpu_utilization", "data_loading"},
		ClusterID:       ineff.ClusterID,
		NodeID:          ineff.NodeID,
	}
}

func (s *Service) genInvestigateStalledTraining(ineff detection.Inefficiency) *recommendation.Recommendation {
	pricing := s.getGPUPricing(ineff.Labels)

	return &recommendation.Recommendation{
		Category:        recommendation.CategoryInvestigateCPU,
		Title:           fmt.Sprintf("Stalled training loop on %s", ineff.GPUID),
		Description:     fmt.Sprintf("GPU %s shows near-zero compute with memory still allocated and processes running. The training loop may be hung.", ineff.GPUID),
		Explanation:     "This pattern typically indicates a deadlock in distributed training, a hung I/O operation, or an infinite loop in preprocessing. Immediate investigation is recommended as the GPU is burning cost without progress.",
		ConfidenceScore: 0.90,
		EstimatedSavings: recommendation.EstimatedSavings{
			CostPerHourUSD:  pricing.CostPerHourUSD,
			CostPerMonthUSD: pricing.CostPerHourUSD * 24, // Waste per day if not fixed
		},
		AffectedMetrics: []string{"gpu_utilization", "gpu_memory_utilization", "running_processes"},
		ClusterID:       ineff.ClusterID,
		NodeID:          ineff.NodeID,
		GPUID:           ineff.GPUID,
		JobID:           ineff.JobID,
	}
}

func (s *Service) genInvestigateThermal(ineff detection.Inefficiency) *recommendation.Recommendation {
	return &recommendation.Recommendation{
		Category:        recommendation.CategoryInvestigateCPU, // Re-using closest category
		Title:           fmt.Sprintf("Critical temperature on %s", ineff.GPUID),
		Description:     fmt.Sprintf("GPU %s temperature peaked at %.1f°C. This risks thermal throttling and hardware damage.", ineff.GPUID, ineff.CurrentValue),
		Explanation:     "Sustained high temperatures cause thermal throttling (reducing performance) and can permanently damage GPU hardware. Check datacenter cooling, GPU fan health, and workload characteristics. Consider reducing power limits or improving airflow.",
		ConfidenceScore: 0.95,
		EstimatedSavings: recommendation.EstimatedSavings{
			PowerWatts: 100, // Potential power savings from reducing thermal load
		},
		AffectedMetrics: []string{"gpu_temperature", "gpu_power_usage"},
		ClusterID:       ineff.ClusterID,
		NodeID:          ineff.NodeID,
		GPUID:           ineff.GPUID,
	}
}

// ============================================================
// HELPERS
// ============================================================

// calculateConfidence computes a confidence score based on how far the current
// value deviates from the threshold. Larger deviations = higher confidence.
func (s *Service) calculateConfidence(currentValue, threshold, baseConfidence float64) float64 {
	if threshold == 0 {
		return baseConfidence
	}
	deviation := math.Abs(currentValue-threshold) / threshold
	confidence := baseConfidence + (1-baseConfidence)*math.Min(deviation, 1.0)
	return math.Min(confidence, 0.99)
}

// getGPUPricing returns pricing for a GPU based on labels.
func (s *Service) getGPUPricing(labels map[string]string) GPUPricing {
	if labels != nil {
		if model, ok := labels["gpu_model"]; ok {
			if p, ok := s.pricing[model]; ok {
				return p
			}
		}
	}
	return s.pricing["default"]
}
