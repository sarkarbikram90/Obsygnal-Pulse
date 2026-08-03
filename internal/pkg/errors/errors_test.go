package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := ErrNotFound.WithDetails("resource ID 123").WithError(fmt.Errorf("underlying DB error"))

	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Contains(t, err.Error(), "RESOURCE_NOT_FOUND")
	assert.Contains(t, err.Error(), "underlying DB error")
	assert.Equal(t, "underlying DB error", err.Unwrap().Error())
}

func TestValidationErrors(t *testing.T) {
	ve := &ValidationErrors{}
	assert.False(t, ve.HasErrors())

	ve.Add("gpu_id", "required")
	assert.True(t, ve.HasErrors())
	assert.Contains(t, ve.Error(), "1 errors")
}

func TestTelemetryError(t *testing.T) {
	inner := fmt.Errorf("socket connection reset")
	te := &TelemetryError{
		Source: "dcgm",
		Metric: "gpu_temp",
		Reason: "timeout",
		Err:    inner,
	}

	assert.Contains(t, te.Error(), "telemetry error [dcgm/gpu_temp]: timeout")
	assert.Equal(t, inner, te.Unwrap())
}

func TestIsRetryable(t *testing.T) {
	assert.True(t, IsRetryable(ErrServiceUnavailable))
	assert.True(t, IsRetryable(ErrTimeout))
	assert.True(t, IsRetryable(ErrRateLimited))
	assert.False(t, IsRetryable(ErrNotFound))
	assert.False(t, IsRetryable(fmt.Errorf("random error")))
}
