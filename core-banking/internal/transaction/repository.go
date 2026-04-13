package transaction

import (
	"context"
	"core-banking/pkg/pagination"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	// Idempotency and Retrieval
	GetIdempotency(ctx context.Context, key string) (*IdempotencyKey, error)
	SaveIdempotency(ctx context.Context, ik *IdempotencyKey) error
	IsTransactionExists(ctx context.Context, partnerRefNo string) (bool, error)
	GetTransactionByReferenceID(ctx context.Context, partnerRefNo string) (*Transaction, error)

	// Account locking + retrieval
	GetSenderForUpdate(ctx context.Context, accountID string) (SenderAccount, error)
	LockReceiver(ctx context.Context, accountID string) (uuid.UUID, error)

	// Write operations
	InsertTransaction(ctx context.Context, p InsertTransactionParams) (string, error)
	InsertLedger(ctx context.Context, p InsertLedgerParams) error
	InsertAccountTransaction(ctx context.Context, p AccountTransaction) error
	DebitAccount(ctx context.Context, accountID string, amount int64) error
	CreditAccount(ctx context.Context, accountID string, amount int64) error
	CompleteTransaction(ctx context.Context, txID string) error

	// Atomic operations
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	WithTx(tx *sqlx.Tx) Repository

	// List
	List(ctx context.Context, f TransactionListFilter) ([]TransactionHistoryResponse, int, *pagination.Cursor, *pagination.Cursor, error)
}

type AtomicTransferParams struct {
	PartnerReferenceID   string
	Amount               int64
	Currency             string
	CustomerID           string
	SourceAccountNo      string
	BeneficiaryAccountNo string
}
