package seeder

import (
	"core-banking/internal/model"

	"github.com/jmoiron/sqlx"
)

func InsertCustomers(tx *sqlx.Tx, data []model.Customer) error {
	query := `
	INSERT INTO customers (
		id, full_name, data_of_birth, nationality,
		email, phone_number, kyc_status, kyc_verified_at,
		risk_level, pep_flag, created_at, updated_at
	)
	VALUES (
		:id, :full_name, :data_of_birth, :nationality,
		:email, :phone_number, :kyc_status, :kyc_verified_at,
		:risk_level, :pep_flag, :created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertAccounts(tx *sqlx.Tx, data []model.Account) error {
	query := `
	INSERT INTO accounts (
		id, customer_id, account_number, account_type,
		currency, status, available_balance, pending_balance,
		overdraft_limit, opened_at, created_at, updated_at
	)
	VALUES (
		:id, :customer_id, :account_number, :account_type,
		:currency, :status, :available_balance, :pending_balance,
		:overdraft_limit, :opened_at, :created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertTransactions(tx *sqlx.Tx, data []model.Transaction) error {
	query := `
	INSERT INTO transactions (
		id, reference_id, external_reference,
		transaction_type, status, amount, currency,
		initiated_by, description, created_at, completed_at
	)
	VALUES (
		:id, :reference_id, :external_reference,
		:transaction_type, :status, :amount, :currency,
		:initiated_by, :description, :created_at, :completed_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertJournalEntries(tx *sqlx.Tx, data []model.JournalEntry) error {
	query := `
	INSERT INTO journal_entries (
		id, transaction_id, journal_type, posted_at, created_at
	)
	VALUES (
		:id, :transaction_id, :journal_type, :posted_at, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertLedgerEntries(tx *sqlx.Tx, data []model.LedgerEntry) error {
	query := `
	INSERT INTO ledger_entries (
		id, journal_id, account_id,
		entry_type, amount, currency,
		balance_after, created_at
	)
	VALUES (
		:id, :journal_id, :account_id,
		:entry_type, :amount, :currency,
		:balance_after, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertPayments(tx *sqlx.Tx, data []model.Payment) error {
	query := `
	INSERT INTO payments (
		id, transaction_id, payment_method,
		provider, status, fee_amount, metadata,
		created_at, updated_at
	)
	VALUES (
		:id, :transaction_id, :payment_method,
		:provider, :status, :fee_amount, :metadata,
		:created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertAuditLogs(tx *sqlx.Tx, data []model.AuditLog) error {
	query := `
	INSERT INTO audit_logs (
		id, actor_id, entity_type, entity_id,
		action, old_value, new_value, ip_address, created_at
	)
	VALUES (
		:id, :actor_id, :entity_type, :entity_id,
		:action, :old_value, :new_value, :ip_address, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertIdempotencyKeys(tx *sqlx.Tx, data []model.IdempotencyKey) error {
	query := `
	INSERT INTO idempotency_keys (
		id, key, request_hash, response, created_at
	)
	VALUES (
		:id, :key, :request_hash, :response, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertFXRates(tx *sqlx.Tx, data []model.FXRate) error {
	query := `
	INSERT INTO fx_rates (
		id, base_currency, quote_currency, rate, effective_at
	)
	VALUES (
		:id, :base_currency, :quote_currency, :rate, :effective_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}
