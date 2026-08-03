package recommendation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecommendationDomain(t *testing.T) {
	now := time.Now()
	rec := Recommendation{
		ID:              "rec-1",
		Category:        CategoryReduceAllocation,
		Status:          StatusActive,
		Title:           "Reduce Allocation",
		ConfidenceScore: 0.95,
		EstimatedSavings: EstimatedSavings{
			CostPerHourUSD:  5.0,
			CostPerMonthUSD: 3600.0,
		},
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	assert.Equal(t, "rec-1", rec.ID)
	assert.Equal(t, StatusActive, rec.Status)
	assert.Equal(t, 0.95, rec.ConfidenceScore)
}
