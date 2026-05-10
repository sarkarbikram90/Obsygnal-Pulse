// Package detection implements the inefficiency detection engine.
//
// Architecture:
//
//	The detection engine follows a rule-based evaluation pattern.
//	Each rule is a pure function: ([]GPUMetric, []NodeMetric) → []Inefficiency
//	Rules are registered in a RuleRegistry and evaluated on a configurable interval.
//
// Why rule-based instead of ML?
//  1. Interpretable: operators can understand exactly why an alert fired
//  2. Tunable: threshold adjustments don't require retraining
//  3. Debuggable: rules produce deterministic output for the same input
//  4. Fast: rule evaluation is O(n) in the number of metrics, no model inference
//  5. Foundation: rule outputs become training data for future ML models
//
// The engine runs as a background worker that:
//  1. Pulls recent metrics from ClickHouse (last evaluation window)
//  2. Evaluates all enabled rules against those metrics
//  3. Deduplicates against currently active inefficiencies
//  4. Persists new inefficiencies to ClickHouse
//  5. Checks if any previously active inefficiencies are now resolved
//  6. Emits events for the alerting and recommendation engines
package detection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// RuleFunc is the signature for a detection rule.
// It receives GPU and node metrics and returns any inefficiencies found.
type RuleFunc func(gpuMetrics []metrics.GPUMetric, nodeMetrics []metrics.NodeMetric) []detection.Inefficiency

// Service orchestrates the inefficiency detection pipeline.
type Service struct {
	metricsRepo ports.MetricsRepository
	ineffRepo   ports.InefficiencyRepository
	cache       ports.CacheStore
	log         *logger.Logger
	cfg         config.DetectionConfig

	rules   []registeredRule
	rulesMu sync.RWMutex

	// Tracking active inefficiencies for dedup
	activeIneffs map[string]*detection.Inefficiency // key: type:gpu_id or type:node_id
	activeMu     sync.RWMutex

	// Listeners get notified when new inefficiencies are detected
	listeners   []func(ctx context.Context, ineff detection.Inefficiency)
	listenersMu sync.RWMutex

	// Stats
	totalEvaluations int64
	totalDetected    int64
	totalResolved    int64
}

type registeredRule struct {
	rule detection.DetectionRule
	fn   RuleFunc
}

// NewService creates a new detection service.
func NewService(
	metricsRepo ports.MetricsRepository,
	ineffRepo ports.InefficiencyRepository,
	cache ports.CacheStore,
	log *logger.Logger,
	cfg config.DetectionConfig,
) *Service {
	s := &Service{
		metricsRepo:  metricsRepo,
		ineffRepo:    ineffRepo,
		cache:        cache,
		log:          log.WithComponent("detection"),
		cfg:          cfg,
		activeIneffs: make(map[string]*detection.Inefficiency),
	}
	s.registerDefaultRules()
	return s
}

// registerDefaultRules configures the built-in detection rules.
// Each rule maps to one InefficiencyType and has a threshold from config.
func (s *Service) registerDefaultRules() {
	s.RegisterRule(detection.DetectionRule{
		Name:        "underutilized_gpu",
		Type:        detection.TypeUnderutilizedGPU,
		Description: "GPU compute utilization below threshold for sustained period",
		Enabled:     true,
		Threshold:   s.cfg.GPUUtilizationThreshold,
		Duration:    5 * time.Minute,
		Severity:    detection.SeverityMedium,
		Cooldown:    15 * time.Minute,
	}, s.ruleUnderutilizedGPU)

	s.RegisterRule(detection.DetectionRule{
		Name:        "idle_gpu",
		Type:        detection.TypeIdleGPU,
		Description: "GPU is completely idle with no running processes",
		Enabled:     true,
		Threshold:   2.0,
		Duration:    10 * time.Minute,
		Severity:    detection.SeverityHigh,
		Cooldown:    15 * time.Minute,
	}, s.ruleIdleGPU)

	s.RegisterRule(detection.DetectionRule{
		Name:        "low_memory_usage",
		Type:        detection.TypeLowMemoryUsage,
		Description: "GPU memory utilization significantly below allocation",
		Enabled:     true,
		Threshold:   s.cfg.MemoryUtilizationThreshold,
		Duration:    10 * time.Minute,
		Severity:    detection.SeverityMedium,
		Cooldown:    30 * time.Minute,
	}, s.ruleLowMemoryUsage)

	s.RegisterRule(detection.DetectionRule{
		Name:        "cpu_starvation",
		Type:        detection.TypeCPUStarvation,
		Description: "Node CPU is saturated while GPUs are underutilized, suggesting CPU bottleneck",
		Enabled:     true,
		Threshold:   90.0,
		Duration:    5 * time.Minute,
		Severity:    detection.SeverityHigh,
		Cooldown:    15 * time.Minute,
	}, s.ruleCPUStarvation)

	s.RegisterRule(detection.DetectionRule{
		Name:        "stalled_training",
		Type:        detection.TypeStalledTraining,
		Description: "GPU utilization dropped dramatically, suggesting stalled training loop",
		Enabled:     true,
		Threshold:   5.0,
		Duration:    15 * time.Minute,
		Severity:    detection.SeverityCritical,
		Cooldown:    30 * time.Minute,
	}, s.ruleStalledTraining)

	s.RegisterRule(detection.DetectionRule{
		Name:        "abnormal_spike",
		Type:        detection.TypeAbnormalSpike,
		Description: "GPU temperature or power exceeding safe operating thresholds",
		Enabled:     true,
		Threshold:   90.0, // Temperature threshold in Celsius
		Duration:    2 * time.Minute,
		Severity:    detection.SeverityCritical,
		Cooldown:    10 * time.Minute,
	}, s.ruleAbnormalSpike)
}

// RegisterRule adds a new detection rule to the engine.
func (s *Service) RegisterRule(rule detection.DetectionRule, fn RuleFunc) {
	s.rulesMu.Lock()
	defer s.rulesMu.Unlock()
	s.rules = append(s.rules, registeredRule{rule: rule, fn: fn})
	s.log.Info().Str("rule", rule.Name).Str("type", string(rule.Type)).Msg("registered detection rule")
}

// OnDetection registers a callback invoked when new inefficiencies are found.
func (s *Service) OnDetection(fn func(ctx context.Context, ineff detection.Inefficiency)) {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()
	s.listeners = append(s.listeners, fn)
}

// Evaluate runs all enabled detection rules against recent metrics.
// This is the core evaluation loop called periodically by Run().
func (s *Service) Evaluate(ctx context.Context) ([]detection.Inefficiency, error) {
	s.totalEvaluations++

	// Pull recent metrics (last evaluation window + buffer)
	window := s.cfg.EvaluationInterval * 2
	now := time.Now()
	from := now.Add(-window)

	gpuMetrics, err := s.metricsRepo.QueryGPUMetrics(ctx, ports.MetricsQuery{
		From:  from,
		To:    now,
		Limit: 10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query GPU metrics: %w", err)
	}

	nodeMetrics, err := s.metricsRepo.QueryNodeMetrics(ctx, ports.MetricsQuery{
		From:  from,
		To:    now,
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query node metrics: %w", err)
	}

	if len(gpuMetrics) == 0 {
		s.log.Debug().Msg("no metrics available for detection evaluation")
		return nil, nil
	}

	// Evaluate all rules
	var allDetected []detection.Inefficiency
	s.rulesMu.RLock()
	rules := s.rules
	s.rulesMu.RUnlock()

	for _, r := range rules {
		if !r.rule.Enabled {
			continue
		}

		detected := r.fn(gpuMetrics, nodeMetrics)
		for i := range detected {
			detected[i].ID = uuid.New().String()
			detected[i].DetectedAt = now
			detected[i].IsActive = true
		}
		allDetected = append(allDetected, detected...)
	}

	// Deduplicate against active inefficiencies
	newIneffs := s.deduplicate(allDetected)

	if len(newIneffs) > 0 {
		// Persist new inefficiencies
		if err := s.ineffRepo.InsertBatch(ctx, newIneffs); err != nil {
			s.log.Error().Err(err).Int("count", len(newIneffs)).Msg("failed to persist inefficiencies")
		} else {
			s.totalDetected += int64(len(newIneffs))
			s.log.Info().Int("new", len(newIneffs)).Int("total_active", len(s.activeIneffs)).Msg("detection evaluation complete")
		}

		// Cache active inefficiencies
		s.cache.Set(ctx, "active_inefficiencies_count", len(s.activeIneffs), 5*time.Minute)

		// Notify listeners
		s.listenersMu.RLock()
		listeners := s.listeners
		s.listenersMu.RUnlock()
		for _, ineff := range newIneffs {
			for _, fn := range listeners {
				fn(ctx, ineff)
			}
		}
	}

	return newIneffs, nil
}

// deduplicate filters out inefficiencies that match already-active ones.
// Key: type + gpu_id (or node_id for node-level detections).
func (s *Service) deduplicate(detected []detection.Inefficiency) []detection.Inefficiency {
	s.activeMu.Lock()
	defer s.activeMu.Unlock()

	var newIneffs []detection.Inefficiency
	for _, d := range detected {
		key := s.inefficiencyKey(d)
		if _, exists := s.activeIneffs[key]; !exists {
			s.activeIneffs[key] = &d
			newIneffs = append(newIneffs, d)
		}
	}
	return newIneffs
}

func (s *Service) inefficiencyKey(d detection.Inefficiency) string {
	target := d.GPUID
	if target == "" {
		target = d.NodeID
	}
	return fmt.Sprintf("%s:%s", d.Type, target)
}

// Run starts the detection engine on a periodic loop.
func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.EvaluationInterval)
	defer ticker.Stop()

	s.log.Info().
		Dur("interval", s.cfg.EvaluationInterval).
		Float64("gpu_threshold", s.cfg.GPUUtilizationThreshold).
		Float64("mem_threshold", s.cfg.MemoryUtilizationThreshold).
		Msg("starting detection engine")

	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("detection engine shutting down")
			return ctx.Err()
		case <-ticker.C:
			if _, err := s.Evaluate(ctx); err != nil {
				s.log.Error().Err(err).Msg("detection evaluation failed")
			}
		}
	}
}

// Stats returns detection engine statistics.
func (s *Service) Stats() map[string]interface{} {
	s.activeMu.RLock()
	activeCount := len(s.activeIneffs)
	s.activeMu.RUnlock()

	return map[string]interface{}{
		"total_evaluations": s.totalEvaluations,
		"total_detected":    s.totalDetected,
		"total_resolved":    s.totalResolved,
		"active_count":      activeCount,
		"rules_count":       len(s.rules),
	}
}

// ============================================================
// DETECTION RULES
// Each rule is a pure function that evaluates metrics and returns
// any detected inefficiencies. Rules are deterministic and testable.
// ============================================================

// ruleUnderutilizedGPU detects GPUs with sustained low compute utilization.
func (s *Service) ruleUnderutilizedGPU(gpuMetrics []metrics.GPUMetric, _ []metrics.NodeMetric) []detection.Inefficiency {
	// Group metrics by GPU ID and compute average utilization
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	var results []detection.Inefficiency

	for gpuID, avg := range gpuAvgs {
		if avg.utilization > 0 && avg.utilization < s.cfg.GPUUtilizationThreshold {
			results = append(results, detection.Inefficiency{
				Type:            detection.TypeUnderutilizedGPU,
				Severity:        detection.SeverityMedium,
				ClusterID:       avg.clusterID,
				NodeID:          avg.nodeID,
				GPUID:           gpuID,
				JobID:           avg.jobID,
				Description:     fmt.Sprintf("GPU %s average utilization is %.1f%% (threshold: %.1f%%)", gpuID, avg.utilization, s.cfg.GPUUtilizationThreshold),
				AffectedMetrics: []string{"gpu_utilization"},
				CurrentValue:    avg.utilization,
				ThresholdValue:  s.cfg.GPUUtilizationThreshold,
			})
		}
	}
	return results
}

// ruleIdleGPU detects GPUs that are completely idle.
func (s *Service) ruleIdleGPU(gpuMetrics []metrics.GPUMetric, _ []metrics.NodeMetric) []detection.Inefficiency {
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	var results []detection.Inefficiency

	for gpuID, avg := range gpuAvgs {
		if avg.utilization < 2.0 && avg.processes == 0 {
			results = append(results, detection.Inefficiency{
				Type:            detection.TypeIdleGPU,
				Severity:        detection.SeverityHigh,
				ClusterID:       avg.clusterID,
				NodeID:          avg.nodeID,
				GPUID:           gpuID,
				Description:     fmt.Sprintf("GPU %s is idle (0%% utilization, no running processes)", gpuID),
				AffectedMetrics: []string{"gpu_utilization", "running_processes"},
				CurrentValue:    avg.utilization,
				ThresholdValue:  2.0,
			})
		}
	}
	return results
}

// ruleLowMemoryUsage detects GPUs with low memory utilization relative to allocation.
func (s *Service) ruleLowMemoryUsage(gpuMetrics []metrics.GPUMetric, _ []metrics.NodeMetric) []detection.Inefficiency {
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	var results []detection.Inefficiency

	for gpuID, avg := range gpuAvgs {
		if avg.utilization > 10 && avg.memoryUtil < s.cfg.MemoryUtilizationThreshold {
			results = append(results, detection.Inefficiency{
				Type:            detection.TypeLowMemoryUsage,
				Severity:        detection.SeverityMedium,
				ClusterID:       avg.clusterID,
				NodeID:          avg.nodeID,
				GPUID:           gpuID,
				JobID:           avg.jobID,
				Description:     fmt.Sprintf("GPU %s memory utilization is %.1f%% — consider smaller GPU or larger batch size", gpuID, avg.memoryUtil),
				AffectedMetrics: []string{"gpu_memory_utilization"},
				CurrentValue:    avg.memoryUtil,
				ThresholdValue:  s.cfg.MemoryUtilizationThreshold,
			})
		}
	}
	return results
}

// ruleCPUStarvation detects nodes where CPU is saturated while GPUs are underutilized.
func (s *Service) ruleCPUStarvation(gpuMetrics []metrics.GPUMetric, nodeMetrics []metrics.NodeMetric) []detection.Inefficiency {
	// Build node CPU usage map
	nodeCPU := make(map[string]float64)
	nodeCluster := make(map[string]string)
	for _, nm := range nodeMetrics {
		nodeCPU[nm.NodeID] = nm.CPUUsage
		nodeCluster[nm.NodeID] = nm.ClusterID
	}

	// Check if high CPU correlates with low GPU utilization
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	nodeGPUAvg := make(map[string]float64)
	nodeGPUCount := make(map[string]int)
	for _, avg := range gpuAvgs {
		nodeGPUAvg[avg.nodeID] += avg.utilization
		nodeGPUCount[avg.nodeID]++
	}

	var results []detection.Inefficiency
	for nodeID, cpuUsage := range nodeCPU {
		if cpuUsage > 90.0 {
			gpuCount := nodeGPUCount[nodeID]
			if gpuCount > 0 {
				avgGPUUtil := nodeGPUAvg[nodeID] / float64(gpuCount)
				if avgGPUUtil < 50.0 {
					results = append(results, detection.Inefficiency{
						Type:            detection.TypeCPUStarvation,
						Severity:        detection.SeverityHigh,
						ClusterID:       nodeCluster[nodeID],
						NodeID:          nodeID,
						Description:     fmt.Sprintf("Node %s CPU at %.1f%% while GPU utilization averages %.1f%% — likely CPU bottleneck", nodeID, cpuUsage, avgGPUUtil),
						AffectedMetrics: []string{"cpu_usage", "gpu_utilization"},
						CurrentValue:    cpuUsage,
						ThresholdValue:  90.0,
					})
				}
			}
		}
	}
	return results
}

// ruleStalledTraining detects GPUs where utilization dropped to near-zero
// while memory is still allocated (indicating a stalled training loop).
func (s *Service) ruleStalledTraining(gpuMetrics []metrics.GPUMetric, _ []metrics.NodeMetric) []detection.Inefficiency {
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	var results []detection.Inefficiency

	for gpuID, avg := range gpuAvgs {
		// Stalled = very low compute but significant memory still allocated
		if avg.utilization < 5.0 && avg.memoryUtil > 30.0 && avg.processes > 0 {
			results = append(results, detection.Inefficiency{
				Type:            detection.TypeStalledTraining,
				Severity:        detection.SeverityCritical,
				ClusterID:       avg.clusterID,
				NodeID:          avg.nodeID,
				GPUID:           gpuID,
				JobID:           avg.jobID,
				Description:     fmt.Sprintf("GPU %s appears stalled: %.1f%% compute utilization but %.1f%% memory still allocated with %d processes", gpuID, avg.utilization, avg.memoryUtil, avg.processes),
				AffectedMetrics: []string{"gpu_utilization", "gpu_memory_utilization", "running_processes"},
				CurrentValue:    avg.utilization,
				ThresholdValue:  5.0,
			})
		}
	}
	return results
}

// ruleAbnormalSpike detects GPUs with dangerous temperature levels.
func (s *Service) ruleAbnormalSpike(gpuMetrics []metrics.GPUMetric, _ []metrics.NodeMetric) []detection.Inefficiency {
	gpuAvgs := s.computeGPUAverages(gpuMetrics)
	var results []detection.Inefficiency

	for gpuID, avg := range gpuAvgs {
		if avg.maxTemp > 90.0 {
			results = append(results, detection.Inefficiency{
				Type:            detection.TypeAbnormalSpike,
				Severity:        detection.SeverityCritical,
				ClusterID:       avg.clusterID,
				NodeID:          avg.nodeID,
				GPUID:           gpuID,
				Description:     fmt.Sprintf("GPU %s temperature peaked at %.1f°C — exceeds safe threshold of 90°C", gpuID, avg.maxTemp),
				AffectedMetrics: []string{"gpu_temperature"},
				CurrentValue:    avg.maxTemp,
				ThresholdValue:  90.0,
			})
		}
	}
	return results
}

// ============================================================
// HELPER: Compute per-GPU averages from a time window of metrics
// ============================================================

type gpuAggregate struct {
	clusterID   string
	nodeID      string
	jobID       string
	utilization float64
	memoryUtil  float64
	maxTemp     float64
	processes   int
	count       int
}

func (s *Service) computeGPUAverages(gpuMetrics []metrics.GPUMetric) map[string]*gpuAggregate {
	aggs := make(map[string]*gpuAggregate)

	for _, m := range gpuMetrics {
		agg, ok := aggs[m.GPUID]
		if !ok {
			agg = &gpuAggregate{
				clusterID: m.ClusterID,
				nodeID:    m.NodeID,
				jobID:     m.JobID,
			}
			aggs[m.GPUID] = agg
		}
		agg.utilization += m.Utilization
		agg.memoryUtil += m.MemoryUtilization
		if m.Temperature > agg.maxTemp {
			agg.maxTemp = m.Temperature
		}
		if m.RunningProcesses > agg.processes {
			agg.processes = m.RunningProcesses
		}
		agg.count++
	}

	// Compute averages
	for _, agg := range aggs {
		if agg.count > 0 {
			agg.utilization /= float64(agg.count)
			agg.memoryUtil /= float64(agg.count)
		}
	}

	return aggs
}
