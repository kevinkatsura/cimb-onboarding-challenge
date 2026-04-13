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
	GetBalance(ctx context.Context, accountID string) (*AccountBalance, error)
	UpdateBalance(ctx context.Context, accountID string, amount int64) error

	// Product
	GetProduct(ctx context.Context, code string) (*Product, error)

	// Customer
	GetCustomerByID(ctx context.Context, id string) (*Customer, error)
	CreateCustomer(ctx context.Context, cust *Customer) error
	UpdateCustomer(ctx context.Context, cust *Customer) error
}
