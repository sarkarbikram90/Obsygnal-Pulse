// Package health provides standardized health check endpoints.
package health

import (
	"context"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents a single health check result.
type Check struct {
	Name    string        `json:"name"`
	Status  Status        `json:"status"`
	Latency time.Duration `json:"latency_ms"`
	Error   string        `json:"error,omitempty"`
}

// Report represents the overall health of the system.
type Report struct {
	Status    Status            `json:"status"`
	Checks   []Check           `json:"checks"`
	Version   string           `json:"version"`
	Uptime    time.Duration    `json:"uptime_seconds"`
	Timestamp time.Time         `json:"timestamp"`
}

// Checker is a function that performs a health check.
type Checker func(ctx context.Context) error

// Registry manages health check registrations.
type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker
	startTime time.Time
	version   string
}

// NewRegistry creates a health check registry.
func NewRegistry(version string) *Registry {
	return &Registry{
		checkers:  make(map[string]Checker),
		startTime: time.Now(),
		version:   version,
	}
}

// Register adds a health check.
func (r *Registry) Register(name string, checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = checker
}

// Check runs all health checks and returns a report.
func (r *Registry) Check(ctx context.Context) *Report {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report := &Report{
		Status:    StatusHealthy,
		Version:   r.version,
		Uptime:    time.Since(r.startTime),
		Timestamp: time.Now(),
	}

	for name, checker := range r.checkers {
		start := time.Now()
		err := checker(ctx)
		latency := time.Since(start)

		check := Check{
			Name:    name,
			Status:  StatusHealthy,
			Latency: latency,
		}

		if err != nil {
			check.Status = StatusUnhealthy
			check.Error = err.Error()
			report.Status = StatusUnhealthy
		} else if latency > 5*time.Second {
			check.Status = StatusDegraded
			if report.Status == StatusHealthy {
				report.Status = StatusDegraded
			}
		}

		report.Checks = append(report.Checks, check)
	}

	return report
}
