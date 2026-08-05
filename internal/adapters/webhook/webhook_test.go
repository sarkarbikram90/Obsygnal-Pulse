package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestWebhookAdapter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	headers := map[string]string{
		"Authorization": "Bearer secret-token",
	}

	adapter := NewAdapter(ts.URL, headers)
	assert.Equal(t, "webhook", adapter.Type())

	notif := ports.Notification{
		Title:    "Custom Webhook Alert",
		Message:  "Test payload",
		Severity: "medium",
	}

	err := adapter.Send(context.Background(), notif)
	assert.NoError(t, err)
}
