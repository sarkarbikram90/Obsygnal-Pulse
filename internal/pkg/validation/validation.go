// Package validation provides reusable input validation helpers.
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// clusterIDPattern allows alphanumeric, hyphens, underscores
	clusterIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
	// emailPattern is a basic email validation
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidateClusterID checks if a cluster ID is valid.
func ValidateClusterID(id string) error {
	if id == "" {
		return fmt.Errorf("cluster_id is required")
	}
	if !clusterIDPattern.MatchString(id) {
		return fmt.Errorf("cluster_id contains invalid characters: %s", id)
	}
	return nil
}

// ValidateEmail checks if an email address is valid.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// ValidatePercentage checks if a value is between 0 and 100.
func ValidatePercentage(name string, value float64) error {
	if value < 0 || value > 100 {
		return fmt.Errorf("%s must be between 0 and 100, got: %f", name, value)
	}
	return nil
}

// ValidateTimeRange checks if from < to and both are reasonable.
func ValidateTimeRange(from, to time.Time) error {
	if from.IsZero() || to.IsZero() {
		return fmt.Errorf("both from and to timestamps are required")
	}
	if from.After(to) {
		return fmt.Errorf("from (%v) must be before to (%v)", from, to)
	}
	if to.Sub(from) > 30*24*time.Hour {
		return fmt.Errorf("time range cannot exceed 30 days")
	}
	return nil
}

// ValidateNonEmpty checks that a string field is non-empty.
func ValidateNonEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidatePositiveInt checks that an integer is positive.
func ValidatePositiveInt(name string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got: %d", name, value)
	}
	return nil
}
