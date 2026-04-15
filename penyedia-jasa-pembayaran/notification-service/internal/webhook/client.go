package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"notification-service/pkg/logging"
	"notification-service/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
)

// WebhookPayload is the payload sent to the partner webhook.
type WebhookPayload struct {
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

// Client sends webhook notifications with retry logic.
type Client struct {
	httpClient *http.Client
	maxRetries int
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		maxRetries: 3,
	}
}

// Result holds webhook delivery result.
type Result struct {
	StatusCode   int
	ResponseBody string
	Err          error
}

// Send delivers a webhook payload to the target URL with exponential backoff retry.
func (c *Client) Send(ctx context.Context, url string, payload WebhookPayload) Result {
	ctx, span := telemetry.Tracer.Start(ctx, "WebhookClient.Send")
	defer span.End()
	span.SetAttributes(
		attribute.String("webhook.url", url),
		attribute.String("webhook.event_type", payload.EventType),
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return Result{Err: fmt.Errorf("marshal failed: %w", err)}
	}

	backoff := 1 * time.Second
	var lastResult Result

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return Result{Err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			logging.Ctx(ctx).Warnw("webhook delivery failed",
				"attempt", attempt, "url", url, "error", err)
			lastResult = Result{Err: err}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		lastResult = Result{StatusCode: resp.StatusCode, ResponseBody: string(respBody)}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logging.Ctx(ctx).Infow("webhook delivered successfully",
				"url", url, "status", resp.StatusCode, "attempt", attempt)
			return lastResult
		}

		logging.Ctx(ctx).Warnw("webhook non-2xx response",
			"attempt", attempt, "status", resp.StatusCode, "url", url)
		time.Sleep(backoff)
		backoff *= 2
	}

	logging.Ctx(ctx).Errorw("webhook delivery exhausted retries",
		"url", url, "last_status", lastResult.StatusCode)
	if lastResult.Err == nil {
		lastResult.Err = fmt.Errorf("webhook delivery failed after %d attempts", c.maxRetries)
	}
	return lastResult
}
