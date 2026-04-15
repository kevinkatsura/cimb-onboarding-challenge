package consumer

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"notification-service/internal/notification"
	"notification-service/internal/webhook"
	"notification-service/pkg/logging"
	"notification-service/pkg/telemetry"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
)

type Handler struct {
	repo      *notification.Repository
	webhook   *webhook.Client
	defaultURL string
}

func NewHandler(repo *notification.Repository, wh *webhook.Client, defaultURL string) *Handler {
	return &Handler{repo: repo, webhook: wh, defaultURL: defaultURL}
}

// ProcessMessage handles a single Kafka message: persist → deliver → update status.
func (h *Handler) ProcessMessage(ctx context.Context, msg kafka.Message) {
	ctx, span := telemetry.Tracer.Start(ctx, "NotificationConsumer.ProcessMessage")
	defer span.End()
	span.SetAttributes(
		attribute.String("kafka.topic", msg.Topic),
		attribute.String("kafka.key", string(msg.Key)),
	)

	logging.Ctx(ctx).Infow("processing kafka message",
		"topic", msg.Topic, "key", string(msg.Key), "partition", msg.Partition)

	// Resolve webhook URL: check event payload for callbackUrl, else default
	callbackURL := h.resolveWebhookURL(msg.Value)

	// Persist notification
	notifID := uuid.New()
	notif := &notification.Notification{
		ID:          notifID,
		EventType:   msg.Topic,
		EventKey:    string(msg.Key),
		Payload:     string(msg.Value),
		CallbackURL: callbackURL,
		Status:      "pending",
		Attempts:    0,
	}
	if err := h.repo.Create(ctx, notif); err != nil {
		logging.Ctx(ctx).Errorw("failed to persist notification", "error", err)
		return
	}

	// Deliver webhook
	var eventData interface{}
	_ = json.Unmarshal(msg.Value, &eventData)

	result := h.webhook.Send(ctx, callbackURL, webhook.WebhookPayload{
		EventType: msg.Topic,
		Data:      eventData,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	// Update status
	status := "sent"
	if result.Err != nil {
		status = "failed"
	}
	statusCode := result.StatusCode
	if err := h.repo.UpdateStatus(ctx, notifID, status, statusCode, result.ResponseBody); err != nil {
		logging.Ctx(ctx).Errorw("failed to update notification status", "error", err)
	}

	logging.Ctx(ctx).Infow("notification processed",
		"id", notifID.String(), "status", status, "webhook_status", statusCode)
}

func (h *Handler) resolveWebhookURL(payload []byte) string {
	var data struct {
		CallbackURL string `json:"callbackUrl"`
	}
	if json.Unmarshal(payload, &data) == nil && data.CallbackURL != "" {
		return data.CallbackURL
	}
	url := os.Getenv("WEBHOOK_URL")
	if url != "" {
		return url
	}
	return h.defaultURL
}
