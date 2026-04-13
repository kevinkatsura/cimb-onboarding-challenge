package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer interface {
	Publish(ctx context.Context, topic string, key string, value interface{}) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewKafkaProducer(brokers []string, logger *zap.Logger) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           10 * time.Millisecond,
		Async:                  false,
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
		MaxAttempts:            25,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	}

	// Manual retry loop for specific transient errors like LeaderNotAvailable
	var lastErr error
	maxRetries := 5
	backoff := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			p.logger.Debug("message published to kafka",
				zap.String("topic", topic),
				zap.Int("attempt", i+1),
			)
			return nil
		}

		lastErr = err
		// Check for specific transient errors (code 3: Unknown Topic, code 5: Leader Not Available)
		if isRetryableError(err) {
			p.logger.Warn("kafka transient error, retrying...",
				zap.Int("attempt", i+1),
				zap.String("topic", topic),
				zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
				continue
			}
		}

		// If it's not a retryable error, or we exhausted retries, break and log
		break
	}

	p.logger.Error("failed to publish message to kafka after retries",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Error(lastErr),
	)
	return lastErr
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := fmt.Sprintf("%v", err)
	if errStr == "Leader Not Available" || errStr == "Unknown Topic Or Partition" {
		return true
	}

	if kerr, ok := err.(kafka.Error); ok {
		return kerr == kafka.LeaderNotAvailable || kerr == kafka.UnknownTopicOrPartition
	}

	return false
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
