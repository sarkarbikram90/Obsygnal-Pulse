package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	log := New("development")
	assert.NotNil(t, log)

	child := log.WithComponent("test").
		WithRequestID("req-123").
		WithTraceID("trace-456").
		WithJobID("job-789").
		WithNodeID("node-01").
		WithGPUID("gpu-01")

	assert.NotNil(t, child)

	prodLog := New("production")
	assert.NotNil(t, prodLog)
}
