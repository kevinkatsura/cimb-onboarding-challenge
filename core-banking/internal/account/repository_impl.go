package account

import (
	"context"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *Account) error {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.Create")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "Create", "Account", "")...)
	span.SetAttributes(telemetry.DBAttrs("postgresql", "banking", "INSERT", "INSERT INTO accounts", -1)...)

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
	RETURNING id, created_at, updated_at, opened_at;`

	return r.DB.QueryRowxContext(
		ctx,
		query,
		acc.CustomerID,
		acc.AccountNumber,
		acc.AccountType,
		acc.Currency,
		acc.OverdraftLimit,
	).StructScan(acc)
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.GetByID")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "GetByID", "Account", id)...)
	span.SetAttributes(telemetry.DBAttrs("postgresql", "banking", "SELECT", "SELECT FROM accounts WHERE id=$1 FOR UPDATE", -1)...)

	var acc Account
	err := r.DB.GetContext(ctx, &acc, `
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
				updated_at,
				deleted_at
		FROM accounts WHERE id=$1 FOR UPDATE;`, id)
	return &acc, err
}

func (r *AccountRepository) List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.List")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "List", "Account", "")...)
	span.SetAttributes(telemetry.DBAttrs("postgresql", "banking", "SELECT", "SELECT FROM accounts (paginated)", -1)...)

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
				updated_at,
				deleted_at
				` + base + `
				` + order + `
				LIMIT $` + fmt.Sprint(idx)
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
			ID:        first.ID.String(),
		}
		nextCursor = &pagination.Cursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID.String(),
		}
	}

	return accounts, total, nextCursor, prevCursor, nil
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.UpdateStatus")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "UpdateStatus", "Account", id)...)
	span.SetAttributes(telemetry.DBAttrs("postgresql", "banking", "UPDATE", "UPDATE accounts SET status", -1)...)

	_, err := r.DB.ExecContext(ctx, `
		UPDATE accounts
		SET status = $1::text,
			updated_at = NOW(),
			closed_at = CASE
				WHEN $1::text = 'closed' AND closed_at IS NULL THEN NOW()
				ELSE closed_at
			END
		WHERE id = $2;`, status, id)
	return err
}

func (r *AccountRepository) SoftDelete(ctx context.Context, id string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.SoftDelete")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "SoftDelete", "Account", id)...)
	span.SetAttributes(telemetry.DBAttrs("postgresql", "banking", "UPDATE", "UPDATE accounts SET deleted_at", -1)...)

	var affectedID string
	err := r.DB.QueryRowxContext(ctx, `
		UPDATE accounts 
		SET deleted_at = NOW(),
			status = 'closed',
			updated_at = NOW(),
			closed_at = COALESCE(closed_at, NOW()) WHERE id = $1 AND deleted_at IS NULL RETURNING id;`, id).Scan(&affectedID)

	if err != nil {
		return err // includes sql.ErrNoRows
	}

	return nil
}
