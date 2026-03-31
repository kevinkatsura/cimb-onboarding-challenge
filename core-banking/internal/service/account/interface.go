package account

import (
	"context"
	"core-banking/internal/dto"

	"core-banking/internal/domain"
)

type Interface interface {
	CreateAccount(ctx context.Context, req dto.CreateAccountRequest) (*domain.Account, error)
	GetAccount(ctx context.Context, id string) (*domain.Account, error)
	ListAccounts(ctx context.Context, f domain.ListFilter) ([]domain.Account, int, string, string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	DeleteAccount(ctx context.Context, id string) error
	UpdateAccountBalance(ctx context.Context, accountID string, amount int64) error
}
