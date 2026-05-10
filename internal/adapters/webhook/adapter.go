// Package webhook implements the Notifier port for generic webhooks.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// Adapter sends notifications via generic HTTP webhooks.
type Adapter struct {
	url        string
	httpClient *http.Client
	headers    map[string]string
}

// NewAdapter creates a new webhook notification adapter.
func NewAdapter(url string, headers map[string]string) *Adapter {
	return &Adapter{
		url:        url,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		headers:    headers,
	}
}

// webhookPayload is the JSON payload sent to the webhook endpoint.
type webhookPayload struct {
	Timestamp time.Time         `json:"timestamp"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Severity  string            `json:"severity"`
	Fields    map[string]string `json:"fields,omitempty"`
	Source    string            `json:"source"`
}

// Send dispatches a notification to the webhook endpoint.
func (a *Adapter) Send(ctx context.Context, notification ports.Notification) error {
	payload := webhookPayload{
		Timestamp: time.Now(),
		Title:     notification.Title,
		Message:   notification.Message,
		Severity:  notification.Severity,
		Fields:    notification.Fields,
		Source:    "obsygnal-pulse",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Type returns the notifier type.
func (a *Adapter) Type() string { return "webhook" }
