package account

import (
	"context"
	"core-banking/pkg/pagination"

	"core-banking/internal/domain"
)

type Repository interface {
	// Account
	Create(ctx context.Context, acc *domain.Account) error
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	List(ctx context.Context, f domain.ListFilter) ([]domain.Account, int, *pagination.Cursor, *pagination.Cursor, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	SoftDelete(ctx context.Context, id string) error
}
