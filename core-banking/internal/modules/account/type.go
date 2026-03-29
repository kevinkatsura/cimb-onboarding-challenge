package account

import (
	"context"
	"core-banking/internal/pkg/pagination"

	"github.com/jmoiron/sqlx"
)

type RepositoryInterface interface {
	Create(tx *sqlx.Tx, acc *Account) error
	GetByID(id string) (*Account, error)
	List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error)
	UpdateStatus(tx *sqlx.Tx, id string, status string) error
	SoftDelete(tx *sqlx.Tx, id string) error
}

type ServiceInterface interface {
	generateAccountNumber() (string, error)
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	DeleteAccount(ctx context.Context, id string) error
}
