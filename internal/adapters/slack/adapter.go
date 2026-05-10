// Package slack implements the Notifier port for Slack webhooks.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/obsygnal/obsygnal-pulse/internal/ports"
)

// Adapter sends notifications via Slack incoming webhooks.
type Adapter struct {
	webhookURL string
	httpClient *http.Client
}

// NewAdapter creates a new Slack notification adapter.
func NewAdapter(webhookURL string) *Adapter {
	return &Adapter{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// slackMessage represents the Slack webhook payload.
type slackMessage struct {
	Text        string        `json:"text"`
	Attachments []attachment  `json:"attachments,omitempty"`
}

type attachment struct {
	Color    string  `json:"color"`
	Title    string  `json:"title"`
	Text     string  `json:"text"`
	Fields   []field `json:"fields,omitempty"`
	Footer   string  `json:"footer"`
	Ts       int64   `json:"ts"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Send dispatches a notification to Slack.
func (a *Adapter) Send(ctx context.Context, notification ports.Notification) error {
	color := "#36a64f" // green
	switch notification.Severity {
	case "warning":
		color = "#ff9900"
	case "critical":
		color = "#ff0000"
	}

	var fields []field
	for k, v := range notification.Fields {
		fields = append(fields, field{Title: k, Value: v, Short: true})
	}

	msg := slackMessage{
		Text: fmt.Sprintf("🚨 *Obsygnal-Pulse Alert*: %s", notification.Title),
		Attachments: []attachment{{
			Color:  color,
			Title:  notification.Title,
			Text:   notification.Message,
			Fields: fields,
			Footer: "Obsygnal-Pulse",
			Ts:     time.Now().Unix(),
		}},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

// Type returns the notifier type.
func (a *Adapter) Type() string { return "slack" }
