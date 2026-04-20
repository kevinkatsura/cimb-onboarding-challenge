package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateCustomer(ctx context.Context, c *Customer) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO account_issuance.customers
		(id, name, email, phone_no, country_code, device_id, device_type, device_model, device_os, onboarding_partner, lang, locale)
		VALUES (:id, :name, :email, :phone_no, :country_code, :device_id, :device_type, :device_model, :device_os, :onboarding_partner, :lang, :locale)`, c)
	return err
}

func (r *PostgresRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	var c Customer
	err := r.db.GetContext(ctx, &c, `SELECT * FROM account_issuance.customers WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, a *Account) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO account_issuance.accounts
		(id, customer_id, account_number, product_code, currency, status)
		VALUES (:id, :customer_id, :account_number, :product_code, :currency, :status)`, a)
	return err
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	var a Account
	err := r.db.GetContext(ctx, &a, `SELECT * FROM account_issuance.accounts WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return &a, nil
}

func (r *PostgresRepository) GetAccountByNumber(ctx context.Context, number string) (*Account, error) {
	var a Account
	err := r.db.GetContext(ctx, &a, `SELECT * FROM account_issuance.accounts WHERE account_number = $1`, number)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return &a, nil
}

func (r *PostgresRepository) CreateBalance(ctx context.Context, b *AccountBalance) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO account_issuance.account_balances
		(account_id, available, pending, currency, version)
		VALUES (:account_id, :available, :pending, :currency, :version)`, b)
	return err
}

func (r *PostgresRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error) {
	var b AccountBalance
	err := r.db.GetContext(ctx, &b, `SELECT * FROM account_issuance.account_balances WHERE account_id = $1`, accountID)
	if err != nil {
		return nil, fmt.Errorf("balance not found: %w", err)
	}
	return &b, nil
}
