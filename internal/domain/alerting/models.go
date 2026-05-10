// Package alerting defines alert domain models.
//
// These models represent alert rules and alert history records.
// They are separate from the ports.Alert struct which defines
// the persistence interface.
package alerting

import "time"

// AlertRule defines a configurable alert rule.
type AlertRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Severity    string         `json:"severity"`
	Condition   AlertCondition `json:"condition"`
	Channels    []string       `json:"channels"` // "slack", "webhook", "email"
	Cooldown    time.Duration  `json:"cooldown"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AlertCondition defines what triggers the alert.
type AlertCondition struct {
	MetricName string        `json:"metric_name"` // e.g., "gpu_utilization"
	Operator   string        `json:"operator"`    // "lt", "gt", "eq"
	Threshold  float64       `json:"threshold"`
	Duration   time.Duration `json:"duration"` // Condition must persist for this duration
}

// Evaluate checks if the condition is met.
func (c *AlertCondition) Evaluate(value float64) bool {
	switch c.Operator {
	case "lt":
		return value < c.Threshold
	case "gt":
		return value > c.Threshold
	case "eq":
		return value == c.Threshold
	case "lte":
		return value <= c.Threshold
	case "gte":
		return value >= c.Threshold
	default:
		return false
	}
}
