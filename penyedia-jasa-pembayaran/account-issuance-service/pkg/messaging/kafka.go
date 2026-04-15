package messaging

import (
	"account-issuance-service/pkg/logging"
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
		Async:                  true,
		AllowAutoTopicCreation: true,
		MaxAttempts:            25,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}
	return &KafkaProducer{writer: writer, logger: logger}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	msg := kafka.Message{Topic: topic, Key: []byte(key), Value: payload}
	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("failed to publish message to kafka", zap.String("topic", topic), zap.Error(err))
		return err
	}
	logging.Logger().Infow("message published to kafka", "topic", topic)
	return nil
}

func (p *KafkaProducer) Close() error { return p.writer.Close() }
