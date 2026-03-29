package account

import (
	"context"
	"core-banking/internal/pkg/pagination"
)

type RepositoryInterface interface {
	Create(acc *Account) error
	GetByID(id string) (*Account, error)
	List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error)
	UpdateStatus(id string, status string) error
	SoftDelete(id string) error
}

type ServiceInterface interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	DeleteAccount(ctx context.Context, id string) error
}
