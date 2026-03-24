package account

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
	CustomerID  *string
	Status      *string
	AccountType *string
	Currency    *string

	Limit     int
	Cursor    *pagination.Cursor
	Direction string // "next" or "prev"
}

func (r *Repository) Create(tx *sqlx.Tx, acc *Account) error {
	query := `
	INSERT INTO accounts(
		id, 
		customer_id, 
		account_number,
		account_type,
		currency,
		status,
		overdraft_limit)
	VALUES(gen_random_uuid(), $1, $2, $3, $4, 'active', $5)
	RETURNING id, created_at, updated_at, opened_at`

	return tx.QueryRowx(
		query,
		acc.CustomerID,
		acc.AccountNumber,
		acc.AccountType,
		acc.Currency,
		acc.OverdraftLimit,
	).StructScan(acc)
}

func (r *Repository) GetByID(id string) (*Account, error) {
	var acc Account
	err := r.DB.Get(&acc, `
		SELECT 	id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at
		FROM accounts WHERE id=$1`, id)
	return &acc, err
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error) {
	var accounts []Account
	var total int

	// Base query
	base := `
		FROM accounts
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	idx := 1

	// Filters
	if f.CustomerID != nil {
		base += fmt.Sprintf(" AND customer_id = $%d", idx)
		args = append(args, *f.CustomerID)
		idx++
	}
	if f.Status != nil {
		base += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *f.Status)
		idx++
	}
	if f.AccountType != nil {
		base += fmt.Sprintf(" AND account_type = $%d", idx)
		args = append(args, *f.AccountType)
		idx++
	}
	if f.Currency != nil {
		base += fmt.Sprintf(" AND currency = $%d", idx)
		args = append(args, *f.Currency)
		idx++
	}

	// Cursor condition
	order := "ORDER BY created_at DESC, id DESC"

	if f.Cursor != nil {
		if f.Direction == "prev" {
			base += fmt.Sprintf(" AND (created_at, id) > ($%d, $%d)", idx, idx+1)
			order = "ORDER BY created_at ASC, id ASC"
		} else {
			// default
			base += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", idx, idx+1)
		}
		args = append(args, f.Cursor.CreatedAt, f.Cursor.ID)
		idx += 2
	}

	// Total count (separate query)
	countQuery := "SELECT COUNT(*) " + base
	err := r.DB.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	// Main query
	query := `
		SELECT	id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at
				` + base + `
				` + order + `
				LIMIT $` + string(idx)
	args = append(args, f.Limit)

	err = r.DB.SelectContext(ctx, &accounts, query, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	// Reverse result if direction=prev
	if f.Direction == "prev" {
		for i, j := 0, len(accounts)-1; i < j; i, j = i+1, j-1 {
			accounts[i], accounts[j] = accounts[j], accounts[i]
		}
	}

	// Build cursor
	var nextCursor, prevCursor *pagination.Cursor

	if len(accounts) > 0 {
		first := accounts[0]
		last := accounts[len(accounts)-1]

		prevCursor = &pagination.Cursor{
			CreatedAt: first.CreatedAt,
			ID:        first.ID,
		}
		nextCursor = &pagination.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return accounts, total, nextCursor, prevCursor, nil
}

func (r *Repository) UpdateStatus(tx *sqlx.Tx, id string, status string) error {
	_, err := tx.Exec(`
		UPDATE accounts
		SET	status=$1,
			updated_at=NOW(),
			closed_at=CASE WHEN $1='closed' THEN NOW() ELSE NULL END
		WHERE id=$2`, status, id)
	return err
}

func (r *Repository) SoftDelete(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(`
		UPDATE accounts
		SET deleted_at=NOW(), status='closed', updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}
