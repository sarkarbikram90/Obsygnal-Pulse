package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracing(t *testing.T) {
	ctx := context.Background()
	ctx, end := StartSpan(ctx, "test_operation", WithAttribute("gpu_id", "gpu-01"))
	assert.NotNil(t, ctx)
	end()

	traceID := TraceID(ctx)
	assert.Equal(t, "", traceID)
}
