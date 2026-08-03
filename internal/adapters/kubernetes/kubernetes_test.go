package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesAdapter(t *testing.T) {
	adapter := NewAdapter("default")
	assert.Equal(t, "default", adapter.namespace)

	pods, err := adapter.ListGPUPods(context.Background())
	assert.Error(t, err)
	assert.Nil(t, pods)

	cap, err := adapter.GetNodeGPUCapacity(context.Background(), "node-01")
	assert.Error(t, err)
	assert.Equal(t, 0, cap)
}
