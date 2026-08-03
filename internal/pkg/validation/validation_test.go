package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	assert.NoError(t, ValidateClusterID("valid-cluster-01"))
	assert.Error(t, ValidateClusterID(""))
	assert.Error(t, ValidateClusterID("invalid cluster@!"))

	assert.NoError(t, ValidateEmail("user@example.com"))
	assert.Error(t, ValidateEmail(""))
	assert.Error(t, ValidateEmail("invalid-email"))

	assert.NoError(t, ValidatePercentage("gpu_util", 50.0))
	assert.Error(t, ValidatePercentage("gpu_util", 150.0))

	now := time.Now()
	assert.NoError(t, ValidateTimeRange(now.Add(-time.Hour), now))
	assert.Error(t, ValidateTimeRange(time.Time{}, now))
	assert.Error(t, ValidateTimeRange(now, now.Add(-time.Hour)))
	assert.Error(t, ValidateTimeRange(now.Add(-31*24*time.Hour), now))

	assert.NoError(t, ValidateNonEmpty("name", "test"))
	assert.Error(t, ValidateNonEmpty("name", "   "))

	assert.NoError(t, ValidatePositiveInt("count", 10))
	assert.Error(t, ValidatePositiveInt("count", -1))
}
