package transaction

import (
	"context"
	"core-banking/internal/pkg/pagination"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{DB: db}
}

type ListFilter struct {
	AccountID *string
	Type      *string
	Status    *string

	Limit     int
	Cursor    *pagination.Cursor
	Direction string
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]TransactionHistoryDTO, int, *pagination.Cursor, *pagination.Cursor, error) {
	var results []TransactionHistoryDTO
	var total int

	base := `
		FROM ledger_entries le
		JOIN journal_entries je ON je.id = le.journal_id
		JOIN transactions tx ON tx.id = je.transaction_id
		WHERE 1=1`

	args := []interface{}{}
	idx := 1

	// Account filter
	if f.AccountID != nil {
		base += fmt.Sprintf(" AND le.account_id = $%d", idx)
		args = append(args, *f.AccountID)
		idx++
	}

	// Transaction filters
	if f.Type != nil {
		base += fmt.Sprintf(" AND t.transaction_type = $%d", idx)
		args = append(args, *f.Type)
		idx++
	}
	if f.Status != nil {
		base += fmt.Sprintf(" AND t.status = $%d", idx)
		args = append(args, *f.Status)
		idx++
	}

	// Cursor (ledger is source of truth)
	order := "ORDER BY le.created_at DESC, le.id DESC"
	if f.Cursor != nil {
		if f.Direction == "prev" {
			base += fmt.Sprintf(" AND (le.created_at, le.id) > ($%d, $%d)", idx, idx+1)
			order = "ORDER BY le.created_at ASC, le.id ASC"
		} else {
			base += fmt.Sprintf(" AND (le.created_at, le.id) < ($%d, $%d)", idx, idx+1)
		}
		args = append(args, f.Cursor.CreatedAt, f.Cursor.ID)
		idx += 2
	}

	// Count
	countQuery := "SELECT COUNT(*) " + base
	if err := r.DB.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, nil, nil, err
	}

	// Main query
	query := `
	SELECT
		le.id AS ledger_entry_id,

		tx.id AS transaction_id,
		tx.reference_id,
		tx.external_reference,
		
		le.account_id,
		
		tx.transaction_type,
		tx.status,
		
		je.journal_type,
		
		le.entry_type,
		le.amount,
		le.currency,
		le.balance_after,
		
		tx.description,
		
		le.created_at,
		tx.completed_at
		` + base + `
		` + order + `
		LIMIT $` + fmt.Sprint(idx)
	args = append(args, f.Limit)

	err := r.DB.SelectContext(ctx, &results, query, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	// Reverse if prev
	if f.Direction == "prev" {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	// Return cursors
	var nextCursor, prevCursor *pagination.Cursor

	if len(results) > 0 {
		first := results[0]
		last := results[len(results)-1]

		prevCursor = &pagination.Cursor{
			CreatedAt: first.CreatedAt,
			ID:        first.LedgerEntryID,
		}
		nextCursor = &pagination.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.LedgerEntryID,
		}
	}

	return results, total, nextCursor, prevCursor, nil
}

func (r *Repository) IsTransactionExists(ctx context.Context, refID string) (bool, error) {
	var exists bool

	err := r.DB.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE reference_id=$1)`,
		refID,
	)

	return exists, err
}

func (r *Repository) GetSenderForUpdate(ctx context.Context, accountID string) (SenderAccount, error) {
	var result SenderAccount

	err := r.DB.GetContext(ctx, &result,
		`SELECT available_balance AS balance, customer_id, account_number
		 FROM accounts
		 WHERE id=$1
		 FOR UPDATE`,
		accountID,
	)

	return result, err
}

func (r *Repository) LockReceiver(ctx context.Context, accountID string) error {
	_, err := r.DB.ExecContext(ctx,
		`SELECT 1 FROM accounts WHERE id=$1 FOR UPDATE`,
		accountID,
	)

	return err
}

func (r *Repository) InsertTransaction(ctx context.Context, p InsertTransactionParams) (string, error) {
	var txID string

	err := r.DB.GetContext(ctx, &txID,
		`INSERT INTO transactions(reference_id, transaction_type, status, amount, currency, initiated_by)
		 VALUES ($1, 'transfer', 'pending', $2, $3, $4)
		 RETURNING id`,
		p.ReferenceID,
		p.Amount,
		p.Currency,
		p.CustomerID,
	)

	return txID, err
}

func (r *Repository) InsertJournal(ctx context.Context, txID string) (string, error) {
	var journalID string

	err := r.DB.GetContext(ctx, &journalID,
		`INSERT INTO journal_entries(transaction_id, journal_type)
		 VALUES ($1, 'transfer')
		 RETURNING id`,
		txID,
	)

	return journalID, err
}

func (r *Repository) InsertLedger(ctx context.Context, p InsertLedgerParams) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO ledger_entries(journal_id, account_id, entry_type, amount, currency)
		 VALUES
		 	($1, $2, 'debit', $3, $4),
		 	($1, $5, 'credit', $3, $4)`,
		p.JournalID,
		p.FromAcc,
		p.Amount,
		p.Currency,
		p.ToAcc,
	)

	return err
}

func (r *Repository) DebitAccount(ctx context.Context, accountID string, amount int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE accounts
		 SET available_balance = available_balance - $1
		 WHERE id=$2`,
		amount,
		accountID,
	)

	return err
}

func (r *Repository) CreditAccount(ctx context.Context, accountID string, amount int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE accounts
		 SET available_balance = available_balance + $1
		 WHERE id=$2`,
		amount,
		accountID,
	)

	return err
}

func (r *Repository) CompleteTransaction(ctx context.Context, txID string) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE transactions
		 SET status='completed', completed_at=NOW()
		 WHERE id=$1`,
		txID,
	)

	return err
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
