package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobMethods(t *testing.T) {
	job := &Job{
		ID:                "job-123",
		Status:            StatusRunning,
		AvgGPUUtilization: 15.0,
		RetryCount:        3,
	}

	assert.True(t, job.IsActive())
	assert.False(t, job.IsFailed())
	assert.True(t, job.IsUnderutilized(30.0))
	assert.True(t, job.HasExcessiveRetries(2))

	job.Status = StatusOOM
	assert.False(t, job.IsActive())
	assert.True(t, job.IsFailed())
}
