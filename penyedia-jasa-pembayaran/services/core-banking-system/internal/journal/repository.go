package journal

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateEntry(ctx context.Context, entry *JournalEntry, lines []JournalLine) error
	GetBalance(ctx context.Context, accountID string) (*AccountLedgerBalance, error)
	UpdateBalance(ctx context.Context, accountID string, delta int64, currency string, entryID uuid.UUID) error
	InitializeBalance(ctx context.Context, accountID, currency string, initialBalance int64) error
}
