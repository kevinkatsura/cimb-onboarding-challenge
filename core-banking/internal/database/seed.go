package database

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Seeder struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Seed(ctx context.Context) error {
	tx, err := BeginSerializableTx(ctx, s.db)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	customers := generateCustomers(50)
	accounts := generateAccounts(customers)

	if err := insertCustomers(ctx, tx, customers); err != nil {
		return err
	}

	if err := insertAccounts(ctx, tx, accounts); err != nil {
		return err
	}

	return tx.Commit()
}

// Customer
type Customer struct {
	ID        uuid.UUID `db:"id"`
	FullName  string    `db:"full_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

func generateCustomers(n int) []Customer {
	rand.Seed(time.Now().UnixNano())

	customers := make([]Customer, 0, n)

	for i := 0; i < n; i++ {
		id := uuid.New()

		customers = append(customers, Customer{
			ID:        id,
			FullName:  fmt.Sprintf("Customer %d", i+1),
			Email:     fmt.Sprintf("customer%d@example.com", i+1),
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(1000)) * time.Hour),
		})
	}

	return customers
}

type Account struct {
	CustomerID uuid.UUID `db:"customer_id"`
	Number     string    `db:"account_number"`
	Balance    int64     `db:"balance"`
	Currency   string    `db:"currency"`
	CreatedAt  time.Time `db:"created_at"`
}

func generateAccounts(customers []Customer) []Account {
	rand.Seed(time.Now().UnixNano())

	accounts := make([]Account, 0, len(customers))

	for i, c := range customers {
		accounts = append(accounts, Account{
			CustomerID: c.ID,
			Number:     fmt.Sprintf("ACC%06d", i+1),
			Balance:    int64(rand.Intn(10_000_000)), // up to 10M
			Currency:   "IDR",
			CreatedAt:  time.Now(),
		})
	}

	return accounts
}

func insertCustomers(ctx context.Context, tx *sqlx.Tx, customers []Customer) error {
	query := `
	INSERT INTO customers (id, full_name, email, created_at)
	VALUES (:id, :full_name, :email, :created_at)
	`

	_, err := tx.NamedExecContext(ctx, query, customers)
	return err
}

func insertAccounts(ctx context.Context, tx *sqlx.Tx, accounts []Account) error {
	query := `
	INSERT INTO accounts (id, customer_id, account_number, balance, currency, created_at)
	VALUES (:id, :customer_id, :account_number, :balance, :currency, :created_at)
	`

	_, err := tx.NamedExecContext(ctx, query, accounts)
	return err
}
