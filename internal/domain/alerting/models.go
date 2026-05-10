// Package alerting defines alert domain models.
package alerting

import "time"

// AlertRule defines a configurable alert rule.
type AlertRule struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Enabled            bool                `json:"enabled"`
	Severity           string              `json:"severity"`
	Condition          AlertCondition      `json:"condition"`
	Channels           []string            `json:"channels"` // "slack", "webhook", "email"
	EscalationPolicyID string              `json:"escalation_policy_id"`
	Cooldown           time.Duration       `json:"cooldown"`
	SuppressionWindows []SuppressionWindow `json:"suppression_windows"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
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

// SuppressionWindow defines periods where alerts are muted (e.g., maintenance).
type SuppressionWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

// EscalationPolicy defines how alerts should be escalated.
type EscalationPolicy struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Steps []EscalationStep `json:"steps"`
}

// EscalationStep is one tier in the escalation chain.
type EscalationStep struct {
	Delay    time.Duration `json:"delay"`
	Channels []string      `json:"channels"` // Slack, Webhook, PagerDuty
}

// AlertGroup groups multiple correlated alerts to prevent storming.
type AlertGroup struct {
	ID        string    `json:"id"`
	ClusterID string    `json:"cluster_id"`
	Severity  string    `json:"severity"`
	AlertIDs  []string  `json:"alert_ids"`
	Status    string    `json:"status"` // "firing", "resolved"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
