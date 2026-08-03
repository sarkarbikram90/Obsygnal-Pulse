package cluster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClusterModel(t *testing.T) {
	now := time.Now()
	c := Cluster{
		ID:        "c1",
		Name:      "prod-cluster",
		Provider:  "gcp",
		Region:    "us-central1",
		CreatedAt: now,
		IsActive:  true,
		Nodes: []Node{
			{
				ID:          "n1",
				Name:        "node-01",
				ClusterID:   "c1",
				MachineType: "a2-highgpu-8g",
				GPUType:     "NVIDIA A100-SXM4-80GB",
				GPUCount:    8,
				CPUCores:    96,
				MemoryGB:    680.0,
				Status:      NodeStatusReady,
			},
		},
	}

	assert.Equal(t, "c1", c.ID)
	assert.Len(t, c.Nodes, 1)
	assert.Equal(t, NodeStatusReady, c.Nodes[0].Status)
}
