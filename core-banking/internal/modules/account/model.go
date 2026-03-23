package account

import "time"

type Account struct {
	ID               string     `db:"id" json:"id"`
	CustomerID       string     `db:"customer_id" json:"customer_id"`
	AccountNumber    string     `db:"account_number" json:"account_number"`
	AccountType      string     `db:"account_type" json:"account_type"`
	Currency         string     `db:"currency" json:"currency"`
	Status           string     `db:"status" json:"status"`
	AvailableBalance int64      `db:"available_balance" json:"available_balance"`
	PendingBalance   int64      `db:"pending_balance" json:"pending_balance"`
	OverdraftLimit   int64      `db:"overdraft_limit" json:"overdraft_limit"`
	OpenedAt         time.Time  `db:"opened_at" json:"opened_at"`
	ClosedAt         *time.Time `db:"closed_at" json:"closed_at"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateAccountRequest struct {
	CustomerID     string `json:"customer_id"`
	AccountType    string `json:"account_type"`
	Currency       string `json:"currency"`
	OverdraftLimit int64  `json:"overdraft_limit"`
}

type UpdateAccountStatusRequest struct {
	Status string `json:"status"`
}
