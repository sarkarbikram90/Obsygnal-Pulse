package alerting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertCondition_Evaluate(t *testing.T) {
	condLT := AlertCondition{Operator: "lt", Threshold: 10.0}
	assert.True(t, condLT.Evaluate(5.0))
	assert.False(t, condLT.Evaluate(15.0))

	condGT := AlertCondition{Operator: "gt", Threshold: 80.0}
	assert.True(t, condGT.Evaluate(90.0))
	assert.False(t, condGT.Evaluate(70.0))

	condEQ := AlertCondition{Operator: "eq", Threshold: 0.0}
	assert.True(t, condEQ.Evaluate(0.0))
	assert.False(t, condEQ.Evaluate(1.0))

	condLTE := AlertCondition{Operator: "lte", Threshold: 10.0}
	assert.True(t, condLTE.Evaluate(10.0))

	condGTE := AlertCondition{Operator: "gte", Threshold: 10.0}
	assert.True(t, condGTE.Evaluate(10.0))

	condUnknown := AlertCondition{Operator: "invalid"}
	assert.False(t, condUnknown.Evaluate(10.0))
}

func TestAlertRule(t *testing.T) {
	now := time.Now()
	rule := AlertRule{
		ID:        "rule-01",
		Name:      "High Temp",
		Enabled:   true,
		Severity:  "high",
		Condition: AlertCondition{Operator: "gt", Threshold: 85.0},
		Channels:  []string{"slack"},
		Cooldown:  5 * time.Minute,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "rule-01", rule.ID)
	assert.True(t, rule.Enabled)
}
