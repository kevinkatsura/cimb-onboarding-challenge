package journal

import (
	"context"
	"fmt"

	"core-banking-system/pkg/logging"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateEntry(ctx context.Context, entry *JournalEntry, lines []JournalLine) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `SET search_path TO ledger`)
	if err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	_, err = tx.NamedExecContext(ctx,
		`INSERT INTO journal_entries (id, transaction_ref, description, entry_date)
		 VALUES (:id, :transaction_ref, :description, :entry_date)`, entry)
	if err != nil {
		return fmt.Errorf("insert journal entry: %w", err)
	}

	for _, line := range lines {
		_, err = tx.NamedExecContext(ctx,
			`INSERT INTO journal_lines (id, journal_entry_id, account_id, debit, credit, currency, balance_after)
			 VALUES (:id, :journal_entry_id, :account_id, :debit, :credit, :currency, :balance_after)`, line)
		if err != nil {
			return fmt.Errorf("insert journal line: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	logging.Ctx(ctx).Infow("journal entry created", "entry_id", entry.ID, "lines", len(lines))
	return nil
}

func (r *PostgresRepository) GetBalance(ctx context.Context, accountID string) (*AccountLedgerBalance, error) {
	var bal AccountLedgerBalance
	err := r.db.GetContext(ctx, &bal,
		`SELECT account_id, current_balance, currency, last_entry_id, updated_at
		 FROM ledger.account_ledger_balances WHERE account_id = $1`, accountID)
	if err != nil {
		return nil, err
	}
	return &bal, nil
}

func (r *PostgresRepository) UpdateBalance(ctx context.Context, accountID string, delta int64, currency string, entryID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ledger.account_ledger_balances (account_id, current_balance, currency, last_entry_id, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (account_id) DO UPDATE SET
		   current_balance = ledger.account_ledger_balances.current_balance + $2,
		   last_entry_id = $4,
		   updated_at = NOW()`,
		accountID, delta, currency, entryID)
	return err
}

func (r *PostgresRepository) InitializeBalance(ctx context.Context, accountID, currency string, initialBalance int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ledger.account_ledger_balances (account_id, current_balance, currency, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (account_id) DO NOTHING`,
		accountID, initialBalance, currency)
	return err
}
