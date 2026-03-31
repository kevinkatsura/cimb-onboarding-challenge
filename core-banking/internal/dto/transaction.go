package dto

import "time"

type TransferRequest struct {
	ReferenceID string `json:"reference_id"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}

type TransferResponse struct {
	Status string `json:"status"`

	TransactionID *string `json:"transaction_id,omitempty"`

	SourceAccount      string `json:"source_account"`
	DestinationAccount string `json:"destination_account"`

	Amount int64 `json:"amount"`

	// Success fields
	SourceBalanceAfter      *int64 `json:"source_balance_after,omitempty"`
	DestinationBalanceAfter *int64 `json:"destination_balance_after,omitempty"`

	// Failure fields
	CurrentBalance *int64 `json:"current_balance,omitempty"`

	Message string `json:"message"`
}

type TransactionHistoryResponse struct {
	LedgerEntryID string  `db:"ledger_entry_id" json:"ledger_entry_id"`
	TransactionID string  `db:"transaction_id" json:"transaction_id"`
	ReferenceID   string  `db:"reference_id" json:"reference_id"`
	ExternalRef   *string `db:"external_reference" json:"external_reference"`

	AccountID string `db:"account_id" json:"account_id"`

	TransactionType string `db:"transaction_type" json:"transaction_type"`
	Status          string `db:"status" json:"status"`

	JournalType *string `db:"journal_type" json:"journal_type"`
	EntryType   *string `db:"entry_type" json:"entry_type"`

	Amount       int64  `db:"amount" json:"amount"`
	Currency     string `db:"currency" json:"currency"`
	BalanceAfter *int64 `db:"balance_after" json:"balance_after"`

	Description *string `db:"description" json:"description"`

	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
}
