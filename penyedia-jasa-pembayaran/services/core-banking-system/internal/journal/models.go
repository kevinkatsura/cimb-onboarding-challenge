package journal

import (
	"time"

	"github.com/google/uuid"
)

type JournalEntry struct {
	ID             uuid.UUID `db:"id"`
	TransactionRef string    `db:"transaction_ref"`
	Description    string    `db:"description"`
	EntryDate      time.Time `db:"entry_date"`
	CreatedAt      time.Time `db:"created_at"`
}

type JournalLine struct {
	ID             uuid.UUID `db:"id"`
	JournalEntryID uuid.UUID `db:"journal_entry_id"`
	AccountID      string    `db:"account_id"`
	Debit          int64     `db:"debit"`
	Credit         int64     `db:"credit"`
	Currency       string    `db:"currency"`
	BalanceAfter   int64     `db:"balance_after"`
	CreatedAt      time.Time `db:"created_at"`
}

type AccountLedgerBalance struct {
	AccountID      string     `db:"account_id"`
	CurrentBalance int64      `db:"current_balance"`
	Currency       string     `db:"currency"`
	LastEntryID    *uuid.UUID `db:"last_entry_id"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

// CreateEntryParams is the input for creating a balanced journal entry.
type CreateEntryParams struct {
	TransactionRef string
	Description    string
	Lines          []LineParam
}

type LineParam struct {
	AccountID string
	Debit     int64
	Credit    int64
	Currency  string
}
