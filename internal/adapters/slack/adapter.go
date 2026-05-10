// Package slack implements the Notifier port for Slack webhooks.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
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
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	TitleLink string  `json:"title_link,omitempty"`
	Text      string  `json:"text"`
	Fields    []field `json:"fields,omitempty"`
	Footer    string  `json:"footer"`
	Ts        int64   `json:"ts"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Send dispatches a notification to Slack.
func (a *Adapter) Send(ctx context.Context, notification ports.Notification) error {
	color := "#36a64f" // green default
	if notification.Severity == "critical" || notification.Severity == "high" {
		color = "#ff0000"
	} else if notification.Severity == "warning" {
		color = "#ff9900"
	}

	var fields []field
	for k, v := range notification.Fields {
		fields = append(fields, field{Title: k, Value: v, Short: true})
	}

	// Enrich with calculated metrics
	if notification.EstimatedWaste > 0 {
		fields = append(fields, field{Title: "Est. Waste (USD)", Value: fmt.Sprintf("$%.2f", notification.EstimatedWaste), Short: true})
	}
	if notification.Remediation != "" {
		fields = append(fields, field{Title: "💡 Remediation Guidance", Value: notification.Remediation, Short: false})
	}

	msg := slackMessage{
		Text: fmt.Sprintf("🚨 *Obsygnal-Pulse Alert*: %s", notification.Title),
		Attachments: []attachment{{
			Color:     color,
			Title:     notification.Title,
			TitleLink: notification.DashboardURL, // Include deep link
			Text:      notification.Message,
			Fields:    fields,
			Footer:    "Obsygnal-Pulse Bot",
			Ts:        time.Now().Unix(),
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

	// Add logic to specifically identify rate limits for better observability
	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("slack rate-limited 429")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

// Type returns the notifier type.
func (a *Adapter) Type() string { return "slack" }
