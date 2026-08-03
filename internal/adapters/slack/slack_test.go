package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

func TestSlackAdapter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL)
	assert.Equal(t, "slack", adapter.Type())

	notif := ports.Notification{
		Title:          "Idle GPU Detected",
		Message:        "GPU idle for 20 mins",
		Severity:       "critical",
		EstimatedWaste: 15.50,
		Remediation:    "Release allocation",
		Fields:         map[string]string{"node": "node-01"},
	}

	err := adapter.Send(context.Background(), notif)
	assert.NoError(t, err)
}
