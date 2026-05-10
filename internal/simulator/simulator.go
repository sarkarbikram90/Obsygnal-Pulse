// Package simulator generates fake GPU telemetry for development and testing.
//
// Architecture decision: The simulator runs as a standalone binary that pushes
// metrics directly into the ingestion pipeline (via Redis Streams) AND exposes
// a Prometheus-compatible /metrics endpoint. This allows testing both push and
// pull collection paths without real GPU hardware.
//
// The simulator models realistic GPU behavior patterns:
// - Normal utilization (70-95%) for healthy training jobs
// - Underutilized GPUs (5-20%) to trigger detection rules
// - Idle GPUs (0%) to test idle detection
// - Memory pressure patterns (approaching OOM)
// - Temperature fluctuations correlated with load
// - Failure scenarios (NaN metrics, stale timestamps)
package simulator

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/domain/metrics"
	"github.com/obsygnal/obsygnal-pulse/internal/pkg/logger"
)

// Config holds simulator configuration.
type Config struct {
	ClusterID    string
	NodeCount    int
	GPUsPerNode  int
	JobCount     int
	TickInterval time.Duration
	// Scenarios to simulate
	EnableIdleGPUs       bool
	EnableUnderutilized  bool
	EnableOOMPressure    bool
	EnableFailures       bool
	EnableStaleTelemetry bool
}

// DefaultConfig returns a reasonable default configuration.
func DefaultConfig() Config {
	return Config{
		ClusterID:           "sim-cluster-01",
		NodeCount:           4,
		GPUsPerNode:         4,
		JobCount:            10,
		TickInterval:        15 * time.Second,
		EnableIdleGPUs:      true,
		EnableUnderutilized: true,
		EnableOOMPressure:   true,
		EnableFailures:      true,
	}
}

// GPUState tracks the simulated state of a single GPU.
type GPUState struct {
	NodeID   string
	GPUID    string
	GPUIndex int
	Model    string
	MemoryMB uint64
	Profile  GPUProfile
	JobID    string
	mu       sync.RWMutex
}

// GPUProfile defines the behavior pattern for a simulated GPU.
type GPUProfile string

const (
	ProfileNormal        GPUProfile = "normal"        // 70-95% utilization
	ProfileUnderutilized GPUProfile = "underutilized" // 5-20% utilization
	ProfileIdle          GPUProfile = "idle"          // 0% utilization
	ProfileOOMPressure   GPUProfile = "oom_pressure"  // Memory approaching limit
	ProfileBursty        GPUProfile = "bursty"        // Alternating high/low
	ProfileStalled       GPUProfile = "stalled"       // Was high, now stuck low
)

// Simulator generates fake GPU telemetry data.
type Simulator struct {
	config Config
	gpus   []*GPUState
	log    *logger.Logger
	rng    *rand.Rand
	tick   int64
}

// New creates a new Simulator.
func New(cfg Config, log *logger.Logger) *Simulator {
	s := &Simulator{
		config: cfg,
		log:    log.WithComponent("simulator"),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.initializeGPUs()
	return s
}

// initializeGPUs creates the simulated GPU inventory.
func (s *Simulator) initializeGPUs() {
	gpuModels := []struct {
		name     string
		memoryMB uint64
	}{
		{"NVIDIA A100-SXM4-80GB", 81920},
		{"NVIDIA A100-SXM4-40GB", 40960},
		{"NVIDIA H100-SXM5-80GB", 81920},
		{"NVIDIA V100-SXM2-32GB", 32768},
	}

	for n := 0; n < s.config.NodeCount; n++ {
		nodeID := fmt.Sprintf("gpu-node-%03d", n)
		for g := 0; g < s.config.GPUsPerNode; g++ {
			model := gpuModels[s.rng.Intn(len(gpuModels))]
			profile := s.assignProfile(n, g)

			gpu := &GPUState{
				NodeID:   nodeID,
				GPUID:    fmt.Sprintf("%s-gpu-%d", nodeID, g),
				GPUIndex: g,
				Model:    model.name,
				MemoryMB: model.memoryMB,
				Profile:  profile,
			}

			// Assign jobs to active GPUs
			if profile != ProfileIdle {
				gpu.JobID = fmt.Sprintf("job-%06d", s.rng.Intn(s.config.JobCount))
			}

			s.gpus = append(s.gpus, gpu)
		}
	}

	s.log.Info().
		Int("total_gpus", len(s.gpus)).
		Int("nodes", s.config.NodeCount).
		Msg("initialized simulated GPU inventory")
}

// assignProfile determines the GPU behavior profile based on simulation config.
func (s *Simulator) assignProfile(nodeIdx, gpuIdx int) GPUProfile {
	totalGPUs := s.config.NodeCount * s.config.GPUsPerNode

	// Deterministic distribution of profiles for reproducibility
	gpuNum := nodeIdx*s.config.GPUsPerNode + gpuIdx
	position := float64(gpuNum) / float64(totalGPUs)

	switch {
	case position < 0.5:
		return ProfileNormal
	case position < 0.65 && s.config.EnableUnderutilized:
		return ProfileUnderutilized
	case position < 0.75 && s.config.EnableIdleGPUs:
		return ProfileIdle
	case position < 0.85 && s.config.EnableOOMPressure:
		return ProfileOOMPressure
	case position < 0.92:
		return ProfileBursty
	default:
		return ProfileStalled
	}
}

// GenerateTick produces one tick of telemetry for all GPUs.
func (s *Simulator) GenerateTick(ctx context.Context) *metrics.MetricBatch {
	s.tick++
	now := time.Now()

	batch := &metrics.MetricBatch{
		GPUMetrics:  make([]metrics.GPUMetric, 0, len(s.gpus)),
		NodeMetrics: make([]metrics.NodeMetric, 0, s.config.NodeCount),
		ReceivedAt:  now,
		Source:      "simulator",
	}

	// Generate GPU metrics
	nodeMetrics := make(map[string]*metrics.NodeMetric)
	for _, gpu := range s.gpus {
		metric := s.generateGPUMetric(gpu, now)
		batch.GPUMetrics = append(batch.GPUMetrics, metric)

		// Aggregate node metrics
		if nm, ok := nodeMetrics[gpu.NodeID]; ok {
			nm.GPUCount++
		} else {
			nodeMetrics[gpu.NodeID] = &metrics.NodeMetric{
				Timestamp: now,
				ClusterID: s.config.ClusterID,
				NodeID:    gpu.NodeID,
				GPUCount:  1,
				PodCount:  s.rng.Intn(20) + 5,
			}
		}
	}

	// Finalize node metrics
	for _, nm := range nodeMetrics {
		nm.CPUUsage = s.jitter(45.0, 30.0)
		nm.MemoryUsage = s.jitter(60.0, 20.0)
		nm.DiskUsage = s.jitter(40.0, 15.0)
		nm.NetworkIn = uint64(s.jitter(500000000, 200000000))
		nm.NetworkOut = uint64(s.jitter(300000000, 100000000))
		batch.NodeMetrics = append(batch.NodeMetrics, *nm)
	}

	return batch
}

// generateGPUMetric creates a single GPU metric based on the GPU's profile.
func (s *Simulator) generateGPUMetric(gpu *GPUState, ts time.Time) metrics.GPUMetric {
	gpu.mu.RLock()
	defer gpu.mu.RUnlock()

	var util, memUtil, temp, power float64

	switch gpu.Profile {
	case ProfileNormal:
		util = s.jitter(82.0, 12.0)
		memUtil = s.jitter(70.0, 15.0)
		temp = s.jitter(72.0, 8.0)
		power = s.jitter(280.0, 40.0)

	case ProfileUnderutilized:
		util = s.jitter(12.0, 7.0)
		memUtil = s.jitter(15.0, 8.0)
		temp = s.jitter(45.0, 5.0)
		power = s.jitter(80.0, 20.0)

	case ProfileIdle:
		util = 0
		memUtil = s.jitter(2.0, 1.0)
		temp = s.jitter(35.0, 3.0)
		power = s.jitter(30.0, 5.0)

	case ProfileOOMPressure:
		// Memory climbs over time toward the limit
		progress := math.Min(float64(s.tick)/100.0, 0.98)
		util = s.jitter(75.0, 10.0)
		memUtil = 50.0 + progress*48.0 + s.jitter(0, 2.0)
		temp = s.jitter(80.0, 5.0)
		power = s.jitter(320.0, 20.0)

	case ProfileBursty:
		// Alternate between high and low every ~10 ticks
		if (s.tick/10)%2 == 0 {
			util = s.jitter(90.0, 5.0)
			memUtil = s.jitter(80.0, 5.0)
		} else {
			util = s.jitter(10.0, 5.0)
			memUtil = s.jitter(20.0, 5.0)
		}
		temp = s.jitter(65.0, 10.0)
		power = s.jitter(200.0, 80.0)

	case ProfileStalled:
		// Was high utilization, now stuck at near-zero
		if s.tick < 20 {
			util = s.jitter(85.0, 5.0)
		} else {
			util = s.jitter(3.0, 2.0)
		}
		memUtil = s.jitter(60.0, 5.0) // Memory still allocated
		temp = s.jitter(50.0, 5.0)
		power = s.jitter(100.0, 20.0)
	}

	// Clamp values
	util = clamp(util, 0, 100)
	memUtil = clamp(memUtil, 0, 100)
	temp = clamp(temp, 20, 100)
	power = clamp(power, 20, 400)

	memTotal := gpu.MemoryMB * 1024 * 1024 // Convert to bytes
	memUsed := uint64(float64(memTotal) * memUtil / 100.0)

	processes := 0
	if util > 1 {
		processes = s.rng.Intn(3) + 1
	}

	return metrics.GPUMetric{
		ID:                fmt.Sprintf("%s-%d", gpu.GPUID, ts.UnixMilli()),
		Timestamp:         ts,
		ClusterID:         s.config.ClusterID,
		NodeID:            gpu.NodeID,
		GPUID:             gpu.GPUID,
		GPUModel:          gpu.Model,
		GPUIndex:          gpu.GPUIndex,
		Utilization:       util,
		MemoryUtilization: memUtil,
		MemoryUsed:        memUsed,
		MemoryTotal:       memTotal,
		MemoryFree:        memTotal - memUsed,
		Temperature:       temp,
		PowerUsage:        power,
		PowerLimit:        400.0,
		FanSpeed:          clamp(temp*0.8, 0, 100),
		PCIeTxBytes:       uint64(s.jitter(1000000000, 500000000)),
		PCIeRxBytes:       uint64(s.jitter(800000000, 400000000)),
		RunningProcesses:  processes,
		JobID:             gpu.JobID,
		Labels: map[string]string{
			"profile":   string(gpu.Profile),
			"simulated": "true",
			"cluster":   s.config.ClusterID,
		},
	}
}

// jitter adds Gaussian noise to a base value.
func (s *Simulator) jitter(base, spread float64) float64 {
	return base + s.rng.NormFloat64()*spread*0.3
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Run starts the simulator, generating telemetry at the configured interval.
func (s *Simulator) Run(ctx context.Context, publisher func(ctx context.Context, batch *metrics.MetricBatch) error) error {
	ticker := time.NewTicker(s.config.TickInterval)
	defer ticker.Stop()

	s.log.Info().
		Str("cluster_id", s.config.ClusterID).
		Int("gpu_count", len(s.gpus)).
		Dur("interval", s.config.TickInterval).
		Msg("starting GPU telemetry simulator")

	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("simulator shutting down")
			return ctx.Err()
		case <-ticker.C:
			batch := s.GenerateTick(ctx)
			if err := publisher(ctx, batch); err != nil {
				s.log.Error().Err(err).Msg("failed to publish simulated telemetry")
			} else {
				s.log.Debug().
					Int("gpu_metrics", len(batch.GPUMetrics)).
					Int("node_metrics", len(batch.NodeMetrics)).
					Msg("published simulated telemetry batch")
			}
		}
	}
}

// GetGPUStates returns the current GPU states for inspection.
func (s *Simulator) GetGPUStates() []*GPUState {
	return s.gpus
}
