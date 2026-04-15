package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"notification-service/config"
	"notification-service/internal/consumer"
	"notification-service/internal/notification"
	"notification-service/internal/webhook"
	"notification-service/pkg/database"
	"notification-service/pkg/logging"
	"notification-service/pkg/telemetry"

	"github.com/segmentio/kafka-go"
)

const defaultWebhookURL = "https://webhook.site/d8f3c0ee-ada4-4fb7-a8f3-3b82724505ff"

func main() {
	logger, _, err := logging.InitLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	shutdown, err := telemetry.InitProvider(context.Background(), "notification-service")
	if err != nil {
		logging.Logger().Fatalw("failed to init telemetry", "error", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	cfg := config.LoadConfig()
	db := database.NewPostgres(cfg)
	defer db.Close()

	repo := notification.NewRepository(db)
	webhookClient := webhook.NewClient()

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = defaultWebhookURL
	}

	handler := consumer.NewHandler(repo, webhookClient, webhookURL)

	// Kafka consumer
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	topics := []string{"account-created", "transfer-completed"}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokers, ","),
		GroupID:  "notification-service",
		Topic:    topics[0], // Primary topic
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	// Second reader for transfer-completed
	reader2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokers, ","),
		GroupID:  "notification-service",
		Topic:    topics[1],
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Health endpoint
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"status":"ok"}`))
		})
		logging.Logger().Infow("health endpoint starting", "port", ":8080")
		http.ListenAndServe(":8080", mux)
	}()

	// Consumer loops
	go func() {
		logging.Logger().Infow("consuming topic", "topic", topics[0])
		for {
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logging.Logger().Errorw("fetch error", "topic", topics[0], "error", err)
				continue
			}
			handler.ProcessMessage(ctx, msg)
			reader.CommitMessages(ctx, msg)
		}
	}()

	go func() {
		logging.Logger().Infow("consuming topic", "topic", topics[1])
		for {
			msg, err := reader2.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logging.Logger().Errorw("fetch error", "topic", topics[1], "error", err)
				continue
			}
			handler.ProcessMessage(ctx, msg)
			reader2.CommitMessages(ctx, msg)
		}
	}()

	logging.Logger().Infow("notification service started", "topics", topics)
	<-ctx.Done()

	logging.Logger().Infow("shutting down")
	reader.Close()
	reader2.Close()
	logging.Logger().Infow("notification service stopped")
}
