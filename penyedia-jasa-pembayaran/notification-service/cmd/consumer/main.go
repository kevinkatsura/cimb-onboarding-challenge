package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"notification-service/config"
	"notification-service/pkg/database"
	"notification-service/pkg/logging"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

const defaultWebhookURL = "https://webhook.site/d8f3c0ee-ada4-4fb7-a8f3-3b82724505ff"

func main() {
	logger, _, _ := logging.InitLogger()
	defer logger.Sync()

	cfg := config.LoadConfig()
	db := database.NewPostgres(cfg)
	defer db.Close()

	// Webhook URL
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = defaultWebhookURL
	}

	// Kafka reader
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	// Kafka topics mapping
	topicsEnv := os.Getenv("KAFKA_TOPICS")
	if topicsEnv == "" {
		topicsEnv = "account_created_v1,transfer_completed_v1"
	}
	topics := strings.Split(topicsEnv, ",")

	// Kafka group ID
	groupID := os.Getenv("KAFKA_CONSUMER_GROUP")
	if groupID == "" {
		groupID = "notification-service-group"
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:      dialer,
		Brokers:     strings.Split(brokers, ","),
		GroupID:     groupID,
		GroupTopics: topics,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		Logger:      kafka.LoggerFunc(logging.Logger().Debugf),
		StartOffset: kafka.FirstOffset,
		ErrorLogger: kafka.LoggerFunc(logging.Logger().Errorf),
	})
	defer reader.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{Timeout: 3 * time.Second}

	logging.Logger().Infow("notification service started", "topics", topics, "webhook", webhookURL)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			logging.Logger().Errorw("fetch error", "error", err)
			continue
		}

		logging.Logger().Infow("event received",
			"topic", msg.Topic,
			"key", string(msg.Key),
			"partition", msg.Partition,
			"offset", msg.Offset,
			"payload", string(msg.Value),
		)

		// 1. Store in DB
		id := uuid.New()
		_, err = db.ExecContext(ctx,
			`INSERT INTO notification.events (id, topic, event_key, payload, webhook_url)
			 VALUES ($1, $2, $3, $4, $5)`,
			id, msg.Topic, string(msg.Key), string(msg.Value), webhookURL)
		if err != nil {
			logging.Logger().Errorw("db insert failed, will retry", "error", err)
			continue
		}

		// 2. POST to webhook
		webhookData := map[string]interface{}{
			"eventType": msg.Topic,
			"data":      json.RawMessage(msg.Value),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		webhookPayload, _ := json.Marshal(webhookData)

		// Verbose log outgoing webhook
		logging.Logger().Infow("delivering webhook", "id", id.String(), "url", webhookURL, "payload", string(webhookPayload))

		resp, err := httpClient.Post(webhookURL, "application/json", bytes.NewReader(webhookPayload))
		status := "sent"
		httpCode := 0
		if err != nil {
			logging.Logger().Warnw("webhook delivery failed", "error", err)
			status = "failed"
		} else {
			httpCode = resp.StatusCode
			resp.Body.Close()
			if httpCode >= 400 {
				status = "failed"
			}
		}

		// 3. Update status
		_, _ = db.ExecContext(ctx,
			`UPDATE notification.events SET status = $1, http_code = $2 WHERE id = $3`,
			status, httpCode, id)

		logging.Logger().Infow("event processed",
			"id", id.String(), "topic", msg.Topic, "status", status, "http_code", httpCode)

		// 4. Commit offset — marks as read
		if err := reader.CommitMessages(ctx, msg); err != nil {
			logging.Logger().Errorw("commit failed", "error", err)
		}
	}

	logging.Logger().Infow("notification service stopped")
}
