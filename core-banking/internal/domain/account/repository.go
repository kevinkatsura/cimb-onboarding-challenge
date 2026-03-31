package account

import (
	"context"
	"core-banking/pkg/pagination"

	"core-banking/internal/domain"
)

type Repository interface {
	// Account
	Create(acc *domain.Account) error
	GetByID(id string) (*domain.Account, error)
	List(ctx context.Context, f domain.ListFilter) ([]domain.Account, int, *pagination.Cursor, *pagination.Cursor, error)
	UpdateStatus(id string, status string) error
	SoftDelete(id string) error
}
