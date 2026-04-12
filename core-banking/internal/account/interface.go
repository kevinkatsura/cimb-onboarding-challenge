package account

import (
	"context"
)

type Interface interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	DeleteAccount(ctx context.Context, id string) error
	UpdateAccountBalance(ctx context.Context, accountID string, amount int64) error
}
