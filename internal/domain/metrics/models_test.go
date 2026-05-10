package metrics

import (
	"testing"
	"time"
)

func TestGPUMetric_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metric  GPUMetric
		wantErr bool
	}{
		{
			name: "valid metric",
			metric: GPUMetric{
				ClusterID:         "cluster-01",
				NodeID:            "node-01",
				GPUID:             "gpu-01",
				Timestamp:         time.Now(),
				Utilization:       85.5,
				MemoryUtilization: 70.0,
				Temperature:       72.0,
			},
			wantErr: false,
		},
		{
			name: "missing cluster_id",
			metric: GPUMetric{
				NodeID:    "node-01",
				GPUID:     "gpu-01",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing node_id",
			metric: GPUMetric{
				ClusterID: "cluster-01",
				GPUID:     "gpu-01",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing gpu_id",
			metric: GPUMetric{
				ClusterID: "cluster-01",
				NodeID:    "node-01",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			metric: GPUMetric{
				ClusterID: "cluster-01",
				NodeID:    "node-01",
				GPUID:     "gpu-01",
			},
			wantErr: true,
		},
		{
			name: "utilization out of range (>100)",
			metric: GPUMetric{
				ClusterID:   "cluster-01",
				NodeID:      "node-01",
				GPUID:       "gpu-01",
				Timestamp:   time.Now(),
				Utilization: 150.0,
			},
			wantErr: true,
		},
		{
			name: "utilization out of range (<0)",
			metric: GPUMetric{
				ClusterID:   "cluster-01",
				NodeID:      "node-01",
				GPUID:       "gpu-01",
				Timestamp:   time.Now(),
				Utilization: -5.0,
			},
			wantErr: true,
		},
		{
			name: "memory utilization out of range",
			metric: GPUMetric{
				ClusterID:         "cluster-01",
				NodeID:            "node-01",
				GPUID:             "gpu-01",
				Timestamp:         time.Now(),
				Utilization:       50.0,
				MemoryUtilization: 120.0,
			},
			wantErr: true,
		},
		{
			name: "temperature out of range",
			metric: GPUMetric{
				ClusterID:         "cluster-01",
				NodeID:            "node-01",
				GPUID:             "gpu-01",
				Timestamp:         time.Now(),
				Utilization:       50.0,
				MemoryUtilization: 50.0,
				Temperature:       200.0,
			},
			wantErr: true,
		},
		{
			name: "boundary value: utilization at 0",
			metric: GPUMetric{
				ClusterID:   "cluster-01",
				NodeID:      "node-01",
				GPUID:       "gpu-01",
				Timestamp:   time.Now(),
				Utilization: 0.0,
				Temperature: 30.0,
			},
			wantErr: false,
		},
		{
			name: "boundary value: utilization at 100",
			metric: GPUMetric{
				ClusterID:   "cluster-01",
				NodeID:      "node-01",
				GPUID:       "gpu-01",
				Timestamp:   time.Now(),
				Utilization: 100.0,
				Temperature: 30.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGPUMetric_IsIdle(t *testing.T) {
	tests := []struct {
		name      string
		metric    GPUMetric
		threshold float64
		want      bool
	}{
		{
			name:      "idle GPU (0% utilization, no processes)",
			metric:    GPUMetric{Utilization: 0, RunningProcesses: 0},
			threshold: 5.0,
			want:      true,
		},
		{
			name:      "low utilization but running processes",
			metric:    GPUMetric{Utilization: 2.0, RunningProcesses: 1},
			threshold: 5.0,
			want:      false,
		},
		{
			name:      "above threshold",
			metric:    GPUMetric{Utilization: 50.0, RunningProcesses: 2},
			threshold: 5.0,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.IsIdle(tt.threshold); got != tt.want {
				t.Errorf("IsIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGPUMetric_MemoryUsagePercent(t *testing.T) {
	tests := []struct {
		name   string
		metric GPUMetric
		want   float64
	}{
		{
			name:   "50% usage",
			metric: GPUMetric{MemoryUsed: 500, MemoryTotal: 1000},
			want:   50.0,
		},
		{
			name:   "zero total memory",
			metric: GPUMetric{MemoryUsed: 100, MemoryTotal: 0},
			want:   0.0,
		},
		{
			name:   "full usage",
			metric: GPUMetric{MemoryUsed: 1000, MemoryTotal: 1000},
			want:   100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.MemoryUsagePercent(); got != tt.want {
				t.Errorf("MemoryUsagePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricBatch_Size(t *testing.T) {
	batch := MetricBatch{
		GPUMetrics:  make([]GPUMetric, 5),
		NodeMetrics: make([]NodeMetric, 3),
	}

	if got := batch.Size(); got != 8 {
		t.Errorf("Size() = %v, want 8", got)
	}
}
