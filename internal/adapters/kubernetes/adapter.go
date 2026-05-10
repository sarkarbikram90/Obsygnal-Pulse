// Package kubernetes implements Kubernetes-aware adapters for pod and node discovery.
//
// This adapter reads pod labels and GPU resource requests from the Kubernetes API
// to enrich telemetry data with job context (which job is running on which GPU).
package kubernetes

import (
	"context"
	"fmt"
)

// Adapter provides Kubernetes cluster introspection.
type Adapter struct {
	// kubeClient *k8s.Clientset — will be initialized in production
	namespace string
}

// NewAdapter creates a new Kubernetes adapter.
func NewAdapter(namespace string) *Adapter {
	return &Adapter{namespace: namespace}
}

// GPUPodInfo represents a pod using GPU resources.
type GPUPodInfo struct {
	PodName    string
	Namespace  string
	NodeName   string
	JobID      string
	GPUCount   int
	Labels     map[string]string
}

// ListGPUPods returns all pods with GPU resource requests in the namespace.
func (a *Adapter) ListGPUPods(ctx context.Context) ([]GPUPodInfo, error) {
	// TODO: Implement using k8s.io/client-go
	return nil, fmt.Errorf("kubernetes adapter not yet implemented")
}

// GetNodeGPUCapacity returns the total GPU capacity for a node.
func (a *Adapter) GetNodeGPUCapacity(ctx context.Context, nodeName string) (int, error) {
	// TODO: Read node.status.capacity["nvidia.com/gpu"]
	return 0, fmt.Errorf("kubernetes adapter not yet implemented")
}
