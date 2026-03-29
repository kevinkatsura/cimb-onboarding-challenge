package transaction

import (
	"context"
	"core-banking/internal/pkg/pagination"
)

type RepositoryInterface interface {
	// Idempotency
	IsTransactionExists(ctx context.Context, refID string) (bool, error)

	// Account locking + retrieval
	GetSenderForUpdate(ctx context.Context, accountID string) (SenderAccount, error)
	LockReceiver(ctx context.Context, accountID string) error

	// Write operations
	InsertTransaction(ctx context.Context, req InsertTransactionParams) (string, error)
	InsertJournal(ctx context.Context, txID string) (string, error)
	InsertLedger(ctx context.Context, p InsertLedgerParams) error
	DebitAccount(ctx context.Context, accountID string, amount int64) error
	CreditAccount(ctx context.Context, accountID string, amount int64) error
	CompleteTransaction(ctx context.Context, txID string) error

	// List
	List(ctx context.Context, f ListFilter) ([]TransactionHistoryDTO, int, *pagination.Cursor, *pagination.Cursor, error)
}
