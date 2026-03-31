package domain

import "core-banking/pkg/pagination"

type ListFilter struct {
	CustomerID  *string
	Status      *string
	AccountType *string
	Currency    *string

	Limit     int
	Cursor    *pagination.Cursor
	Direction string // "next" or "prev"
}

type TransactionListFilter struct {
	AccountID *string
	Limit     int
	Cursor    *pagination.Cursor
	Direction string
	Type      *string
	Status    *string
}

type SenderAccount struct {
	Balance    int64  `db:"balance"`
	CustomerID string `db:"customer_id"`
	AccountNo  string `db:"account_number"`
}

type InsertTransactionParams struct {
	ReferenceID string
	Amount      int64
	Currency    string
	CustomerID  string
}

type InsertLedgerParams struct {
	JournalID string
	FromAcc   string
	ToAcc     string
	Amount    int64
	Currency  string
}
