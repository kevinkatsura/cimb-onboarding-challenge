package repository

import (
	"context"
	"time"
)

type Account struct {
	AccountNumber string    `json:"account_number"`
	AccountID     string    `json:"account_id"`
	CustomerID    string    `json:"customer_id"`
	ProductCode   string    `json:"product_code"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Balance       int64     `json:"balance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                       int64     `json:"id"`
	TransactionRef           string    `json:"transaction_ref"`
	SourceAccountNumber      string    `json:"source_account_number"`
	BeneficiaryAccountNumber string    `json:"beneficiary_account_number"`
	Amount                   int64     `json:"amount"`
	Currency                 string    `json:"currency"`
	CreatedAt                time.Time `json:"created_at"`
}

func (db *PostgresDatabase) UpsertAccount(ctx context.Context, acc Account) error {
	query := `
		INSERT INTO ais.accounts (account_number, account_id, customer_id, product_code, currency, status, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (account_number) DO UPDATE SET
			status = EXCLUDED.status,
			balance = EXCLUDED.balance,
			updated_at = NOW()
	`
	_, err := db.Pool.Exec(ctx, query, acc.AccountNumber, acc.AccountID, acc.CustomerID, acc.ProductCode, acc.Currency, acc.Status, acc.Balance, time.Now(), time.Now())
	return err
}

func (db *PostgresDatabase) UpsertTransaction(ctx context.Context, tx Transaction) error {
	query := `
		INSERT INTO ais.transactions (transaction_ref, source_account_number, beneficiary_account_number, amount, currency, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (transaction_ref) DO NOTHING
	`
	_, err := db.Pool.Exec(ctx, query, tx.TransactionRef, tx.SourceAccountNumber, tx.BeneficiaryAccountNumber, tx.Amount, tx.Currency, time.Now())
	return err
}

func (db *PostgresDatabase) UpdateBalance(ctx context.Context, accountNo string, amountDelta int64) error {
	query := `
		UPDATE ais.accounts
		SET balance = balance + $1, updated_at = NOW()
		WHERE account_number = $2
	`
	_, err := db.Pool.Exec(ctx, query, amountDelta, accountNo)
	return err
}

func (db *PostgresDatabase) GetBalance(ctx context.Context, accountNo string) (int64, string, error) {
	var balance int64
	var currency string
	err := db.Pool.QueryRow(ctx, "SELECT balance, currency FROM ais.accounts WHERE account_number = $1", accountNo).Scan(&balance, &currency)
	return balance, currency, err
}

func (db *PostgresDatabase) GetAccountInfo(ctx context.Context, accountNo string) (Account, error) {
	var acc Account
	err := db.Pool.QueryRow(ctx, "SELECT account_number, account_id, customer_id, product_code, currency, status, balance, created_at, updated_at FROM ais.accounts WHERE account_number = $1", accountNo).
		Scan(&acc.AccountNumber, &acc.AccountID, &acc.CustomerID, &acc.ProductCode, &acc.Currency, &acc.Status, &acc.Balance, &acc.CreatedAt, &acc.UpdatedAt)
	return acc, err
}

func (db *PostgresDatabase) GetLastTransactionAsSource(ctx context.Context, accountNo string) (Transaction, error) {
	var tx Transaction
	err := db.Pool.QueryRow(ctx, "SELECT transaction_ref, source_account_number, beneficiary_account_number, amount, currency, created_at FROM ais.transactions WHERE source_account_number = $1 ORDER BY created_at DESC LIMIT 1", accountNo).
		Scan(&tx.TransactionRef, &tx.SourceAccountNumber, &tx.BeneficiaryAccountNumber, &tx.Amount, &tx.Currency, &tx.CreatedAt)
	return tx, err
}

func (db *PostgresDatabase) GetLastTransactionAsBeneficiary(ctx context.Context, accountNo string) (Transaction, error) {
	var tx Transaction
	err := db.Pool.QueryRow(ctx, "SELECT transaction_ref, source_account_number, beneficiary_account_number, amount, currency, created_at FROM ais.transactions WHERE beneficiary_account_number = $1 ORDER BY created_at DESC LIMIT 1", accountNo).
		Scan(&tx.TransactionRef, &tx.SourceAccountNumber, &tx.BeneficiaryAccountNumber, &tx.Amount, &tx.Currency, &tx.CreatedAt)
	return tx, err
}

func (db *PostgresDatabase) GetAverageAmountLast30Transactions(ctx context.Context, accountNo string) (float64, string, error) {
	query := `
		WITH last_tx AS (
			SELECT amount, currency
			FROM ais.transactions
			WHERE source_account_number = $1 OR beneficiary_account_number = $1
			ORDER BY created_at DESC
			LIMIT 30
		)
		SELECT COALESCE(AVG(amount), 0), MAX(currency)
		FROM last_tx
	`
	var avg float64
	var currency string
	err := db.Pool.QueryRow(ctx, query, accountNo).Scan(&avg, &currency)
	if currency == "" {
		currency = "IDR" // Default if no tx
	}
	return avg, currency, err
}
