package account

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateCustomer(ctx context.Context, c *Customer) error
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	CreateAccount(ctx context.Context, a *Account) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetAccountByNumber(ctx context.Context, number string) (*Account, error)
	CreateBalance(ctx context.Context, b *AccountBalance) error
	GetBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error)
}
