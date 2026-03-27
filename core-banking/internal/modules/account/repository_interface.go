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
