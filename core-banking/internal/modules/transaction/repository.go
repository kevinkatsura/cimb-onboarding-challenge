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
