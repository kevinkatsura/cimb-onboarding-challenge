package transfer

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

func (r *PostgresRepository) CreateTransaction(ctx context.Context, tx *Transaction) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO payment_initiation.transactions
		(id, partner_reference_no, reference_no, type, status, amount, currency, fee_amount, fee_type, remark, fraud_decision, fraud_event_id)
		VALUES (:id, :partner_reference_no, :reference_no, :type, :status, :amount, :currency, :fee_amount, :fee_type, :remark, :fraud_decision, :fraud_event_id)`, tx)
	return err
}

func (r *PostgresRepository) CreateTransferDetail(ctx context.Context, detail *TransferDetail) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO payment_initiation.transfer_details
		(id, transaction_id, source_account_no, source_account_name, beneficiary_account_no, beneficiary_account_name, beneficiary_email)
		VALUES (:id, :transaction_id, :source_account_no, :source_account_name, :beneficiary_account_no, :beneficiary_account_name, :beneficiary_email)`, detail)
	return err
}

func (r *PostgresRepository) GetTransactionByRef(ctx context.Context, refNo string) (*Transaction, error) {
	var tx Transaction
	err := r.db.GetContext(ctx, &tx,
		`SELECT * FROM payment_initiation.transactions WHERE reference_no = $1`, refNo)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	return &tx, nil
}

func (r *PostgresRepository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payment_initiation.transactions SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}
