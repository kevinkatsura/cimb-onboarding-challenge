package transaction

import (
	"context"
	"core-banking/pkg/pagination"

	"core-banking/internal/domain"
	"core-banking/internal/dto"
)

type Repository interface {
	// Idempotency
	IsTransactionExists(ctx context.Context, refID string) (bool, error)

	// Account locking + retrieval
	GetSenderForUpdate(ctx context.Context, accountID string) (domain.SenderAccount, error)
	LockReceiver(ctx context.Context, accountID string) error

	// Write operations
	InsertTransaction(ctx context.Context, req domain.InsertTransactionParams) (string, error)
	InsertJournal(ctx context.Context, txID string) (string, error)
	InsertLedger(ctx context.Context, p domain.InsertLedgerParams) error
	DebitAccount(ctx context.Context, accountID string, amount int64) error
	CreditAccount(ctx context.Context, accountID string, amount int64) error
	CompleteTransaction(ctx context.Context, txID string) error

	// List
	List(ctx context.Context, f domain.TransactionListFilter) ([]dto.TransactionHistoryResponse, int, *pagination.Cursor, *pagination.Cursor, error)
}
