package transaction

import (
	"context"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Ext interface {
	sqlx.QueryerContext
	sqlx.ExecerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type TransactionRepository struct {
	ext Ext
	db  *sqlx.DB
}

func NewRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{
		ext: db,
		db:  db,
	}
}

func (r *TransactionRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *TransactionRepository) WithTx(tx *sqlx.Tx) Repository {
	return &TransactionRepository{
		ext: tx,
		db:  r.db,
	}
}

func (r *TransactionRepository) List(ctx context.Context, f TransactionListFilter) ([]TransactionHistoryResponse, int, *pagination.Cursor, *pagination.Cursor, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.List")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("TransactionRepository", "List", "Transaction", "")...)

	var results []TransactionHistoryResponse
	var total int

	base := `
		FROM ledger_entries le
		JOIN accounts a ON a.id = le.account_id
		JOIN journals j ON j.id = le.journal_id
		JOIN transactions tx ON tx.id = j.transaction_id
		WHERE 1=1`

	args := []interface{}{}
	idx := 1

	if f.AccountID != nil {
		base += fmt.Sprintf(" AND le.account_id = $%d", idx)
		args = append(args, *f.AccountID)
		idx++
	}

	if f.Type != nil {
		base += fmt.Sprintf(" AND tx.transaction_type = $%d", idx)
		args = append(args, *f.Type)
		idx++
	}
	if f.Status != nil {
		base += fmt.Sprintf(" AND tx.status = $%d", idx)
		args = append(args, *f.Status)
		idx++
	}

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

	countQuery := "SELECT COUNT(*) " + base
	if err := r.ext.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, nil, nil, err
	}

	query := `
	SELECT
		le.id AS ledger_entry_id,
		tx.id AS transaction_id,
		tx.partner_reference_no,
		tx.reference_no,
		le.account_id,
		a.account_number,
		tx.transaction_type,
		tx.status,
		j.journal_type,
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

	err := r.ext.SelectContext(ctx, &results, query, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	if f.Direction == "prev" {
		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

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

func (r *TransactionRepository) IsTransactionExists(ctx context.Context, refID string) (bool, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.IsTransactionExists")
	defer span.End()

	var exists bool
	err := r.ext.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE partner_reference_no=$1)`,
		refID,
	)
	return exists, err
}

func (r *TransactionRepository) GetTransactionByReferenceID(ctx context.Context, refID string) (*Transaction, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.GetTransactionByReferenceID")
	defer span.End()

	var tx Transaction
	err := r.ext.GetContext(ctx, &tx,
		`SELECT * FROM transactions WHERE partner_reference_no=$1`,
		refID,
	)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) GetSenderForUpdate(ctx context.Context, accountID string) (SenderAccount, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.GetSenderForUpdate")
	defer span.End()

	var result SenderAccount
	err := r.ext.GetContext(ctx, &result,
		`SELECT available_balance AS balance, customer_id, account_number
		 FROM accounts
		 WHERE id=$1
		 FOR UPDATE`,
		accountID,
	)
	return result, err
}

func (r *TransactionRepository) LockReceiver(ctx context.Context, accountID string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.LockReceiver")
	defer span.End()

	_, err := r.ext.ExecContext(ctx,
		`SELECT 1 FROM accounts WHERE id=$1 FOR UPDATE`,
		accountID,
	)
	return err
}

func (r *TransactionRepository) InsertTransaction(ctx context.Context, p InsertTransactionParams) (string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.InsertTransaction")
	defer span.End()

	var txID string
	err := r.ext.GetContext(ctx, &txID, `
		INSERT INTO transactions (
			partner_reference_no, transaction_type, status, amount, currency
		) VALUES ($1, 'transfer', 'pending', $2, $3)
		RETURNING id`,
		p.PartnerReferenceNo, p.Amount, p.Currency,
	)
	if err != nil {
		return "", err
	}

	_, err = r.ext.ExecContext(ctx, `
		INSERT INTO transfer_details (
			transaction_id, source_account_no, beneficiary_account_no, beneficiary_account_name,
			beneficiary_address, beneficiary_bank_code, beneficiary_bank_name,
			beneficiary_email, customer_reference, fee_type, transaction_date,
			remark, originator_infos, additional_info
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		txID, p.SourceAccountNo, p.BeneficiaryAccountNo, p.BeneficiaryAccountName,
		p.BeneficiaryAddress, p.BeneficiaryBankCode, p.BeneficiaryBankName,
		p.BeneficiaryEmail, p.CustomerReference, p.FeeType, p.TransactionDate,
		p.Remark, p.OriginatorInfos, p.AdditionalInfo,
	)

	return txID, err
}

func (r *TransactionRepository) InsertJournal(ctx context.Context, txID string) (string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.InsertJournal")
	defer span.End()

	var journalID string
	err := r.ext.GetContext(ctx, &journalID,
		`INSERT INTO journals (transaction_id, journal_type, status)
		 VALUES ($1, 'transfer', 'posted')
		 RETURNING id`,
		txID,
	)
	return journalID, err
}

func (r *TransactionRepository) InsertLedger(ctx context.Context, p InsertLedgerParams) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.InsertLedger")
	defer span.End()

	if len(p.Entries) == 0 {
		return nil
	}

	query := `INSERT INTO ledger_entries(journal_id, account_id, entry_type, amount, currency) VALUES `
	vals := []interface{}{}
	for i, e := range p.Entries {
		pos := i * 5
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d),", pos+1, pos+2, pos+3, pos+4, pos+5)
		vals = append(vals, p.JournalID, e.AccountID, e.EntryType, e.Amount, e.Currency)
	}
	query = query[:len(query)-1] // Remove trailing comma

	_, err := r.ext.ExecContext(ctx, query, vals...)
	return err
}

func (r *TransactionRepository) DebitAccount(ctx context.Context, accountID string, amount int64) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.DebitAccount")
	defer span.End()

	_, err := r.ext.ExecContext(ctx,
		`UPDATE accounts
		 SET available_balance = available_balance - $1
		 WHERE id=$2`,
		amount,
		accountID,
	)
	return err
}

func (r *TransactionRepository) CreditAccount(ctx context.Context, accountID string, amount int64) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.CreditAccount")
	defer span.End()

	_, err := r.ext.ExecContext(ctx,
		`UPDATE accounts
		 SET available_balance = available_balance + $1
		 WHERE id=$2`,
		amount,
		accountID,
	)
	return err
}

func (r *TransactionRepository) CompleteTransaction(ctx context.Context, txID string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.CompleteTransaction")
	defer span.End()

	_, err := r.ext.ExecContext(ctx,
		`UPDATE transactions
		 SET status='completed', completed_at=NOW(), updated_at=NOW(), reference_no=id::text
		 WHERE id=$1`,
		txID,
	)
	return err
}

func (r *TransactionRepository) GetIdempotency(ctx context.Context, key string) (*IdempotencyKey, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.GetIdempotency")
	defer span.End()

	var ik IdempotencyKey
	err := r.ext.GetContext(ctx, &ik, "SELECT * FROM idempotency_keys WHERE key = $1", key)
	if err != nil {
		return nil, err
	}
	return &ik, nil
}

func (r *TransactionRepository) SaveIdempotency(ctx context.Context, ik *IdempotencyKey) error {
	ctx, span := telemetry.Tracer.Start(ctx, "TransactionRepository.SaveIdempotency")
	defer span.End()

	_, err := r.ext.NamedExecContext(ctx, `
		INSERT INTO idempotency_keys (id, key, request_hash, response_code, response_message, response_body, created_at)
		VALUES (:id, :key, :request_hash, :response_code, :response_message, :response_body, :created_at)
		ON CONFLICT (key) DO UPDATE SET
			response_code = EXCLUDED.response_code,
			response_message = EXCLUDED.response_message,
			response_body = EXCLUDED.response_body
	`, ik)
	return err
}
