package account

import (
	"context"
	"core-banking/pkg/pagination"
)

type Repository interface {
	// Account
	Create(ctx context.Context, acc *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	SoftDelete(ctx context.Context, id string) error
}
