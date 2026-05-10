package simulator

import (
	"context"
	"testing"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	log := logger.New("development")

	sim := New(cfg, log)

	expectedGPUs := cfg.NodeCount * cfg.GPUsPerNode
	if len(sim.gpus) != expectedGPUs {
		t.Errorf("expected %d GPUs, got %d", expectedGPUs, len(sim.gpus))
	}
}

func TestGenerateTick(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeCount = 2
	cfg.GPUsPerNode = 2
	log := logger.New("development")

	sim := New(cfg, log)
	batch := sim.GenerateTick(context.Background())

	if len(batch.GPUMetrics) != 4 {
		t.Errorf("expected 4 GPU metrics, got %d", len(batch.GPUMetrics))
	}

	if len(batch.NodeMetrics) != 2 {
		t.Errorf("expected 2 node metrics, got %d", len(batch.NodeMetrics))
	}

	if batch.Source != "simulator" {
		t.Errorf("expected source 'simulator', got %s", batch.Source)
	}

	// Validate all metrics
	for _, m := range batch.GPUMetrics {
		if err := m.Validate(); err != nil {
			t.Errorf("generated metric failed validation: %v", err)
		}

		if m.ClusterID != cfg.ClusterID {
			t.Errorf("expected cluster_id %s, got %s", cfg.ClusterID, m.ClusterID)
		}

		if m.Utilization < 0 || m.Utilization > 100 {
			t.Errorf("utilization out of range: %f", m.Utilization)
		}

		if m.Temperature < 20 || m.Temperature > 100 {
			t.Errorf("temperature out of range: %f", m.Temperature)
		}
	}
}

func TestGPUProfiles(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeCount = 10
	cfg.GPUsPerNode = 4
	cfg.EnableIdleGPUs = true
	cfg.EnableUnderutilized = true
	cfg.EnableOOMPressure = true
	log := logger.New("development")

	sim := New(cfg, log)

	profileCounts := make(map[GPUProfile]int)
	for _, gpu := range sim.gpus {
		profileCounts[gpu.Profile]++
	}

	// Should have a mix of profiles
	if profileCounts[ProfileNormal] == 0 {
		t.Error("expected at least one normal profile GPU")
	}

	if profileCounts[ProfileIdle] == 0 {
		t.Error("expected at least one idle profile GPU")
	}
}

func TestMultipleTicks(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeCount = 2
	cfg.GPUsPerNode = 2
	log := logger.New("development")

	sim := New(cfg, log)

	// Generate multiple ticks and verify consistency
	for i := 0; i < 10; i++ {
		batch := sim.GenerateTick(context.Background())
		if batch.Size() == 0 {
			t.Errorf("tick %d produced empty batch", i)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{50, 0, 100, 50},
		{-5, 0, 100, 0},
		{150, 0, 100, 100},
		{0, 0, 100, 0},
		{100, 0, 100, 100},
	}

	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}
