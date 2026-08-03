package health

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthRegistry(t *testing.T) {
	registry := NewRegistry("1.0.0")

	registry.Register("clickhouse", func(ctx context.Context) error {
		return nil
	})

	registry.Register("redis", func(ctx context.Context) error {
		return fmt.Errorf("connection refused")
	})

	report := registry.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, report.Status)
	assert.Equal(t, "1.0.0", report.Version)
	assert.Len(t, report.Checks, 2)
}
