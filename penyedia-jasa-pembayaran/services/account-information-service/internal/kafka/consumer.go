package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"account-information-service/internal/config"
	"account-information-service/internal/repository"
)

type Consumer struct {
	reader *kafka.Reader
	repo   *repository.PostgresDatabase
	logger *zap.SugaredLogger
}

func NewConsumer(cfg config.Config, repo *repository.PostgresDatabase, logger *zap.SugaredLogger) *Consumer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     "account-information-service-group",
		GroupTopics: []string{"account_created_v1", "transfer_completed_v1"},
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	return &Consumer{reader: r, repo: repo, logger: logger}
}

type AccountCreatedEvent struct {
	AccountID     string `json:"account_id"`
	CustomerID    string `json:"customer_id"`
	AccountNumber string `json:"account_number"`
	ProductCode   string `json:"product_code"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}

type TransferCompletedEvent struct {
	TransactionID      string `json:"TransactionID"`
	ReferenceNo        string `json:"ReferenceNo"`
	Amount             int64  `json:"Amount"`
	Currency           string `json:"Currency"`
	SourceAccount      string `json:"SourceAccount"`
	BeneficiaryAccount string `json:"BeneficiaryAccount"`
	Status             string `json:"Status"`
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				c.logger.Errorw("failed to fetch message", "error", err)
				continue
			}

			c.logger.Infow("event received", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)

			switch msg.Topic {
			case "account_created_v1":
				var ev AccountCreatedEvent
				if err := json.Unmarshal(msg.Value, &ev); err == nil {
					acc := repository.Account{
						AccountNumber: ev.AccountNumber,
						AccountID:     ev.AccountID,
						CustomerID:    ev.CustomerID,
						ProductCode:   ev.ProductCode,
						Currency:      ev.Currency,
						Status:        ev.Status,
						Balance:       0,
					}
					c.repo.UpsertAccount(ctx, acc)
				}
			case "transfer_completed_v1":
				var ev TransferCompletedEvent
				if err := json.Unmarshal(msg.Value, &ev); err == nil && ev.Status == "completed" {
					tx := repository.Transaction{
						TransactionRef:           ev.ReferenceNo,
						SourceAccountNumber:      ev.SourceAccount,
						BeneficiaryAccountNumber: ev.BeneficiaryAccount,
						Amount:                   ev.Amount,
						Currency:                 ev.Currency,
					}
					if err := c.repo.UpsertTransaction(ctx, tx); err == nil {
						c.repo.UpdateBalance(ctx, ev.SourceAccount, -ev.Amount)
						c.repo.UpdateBalance(ctx, ev.BeneficiaryAccount, ev.Amount)
					}
				}
			}

			c.reader.CommitMessages(ctx, msg)
		}
	}()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
