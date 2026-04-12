package seeder

import (
	"core-banking/internal/account"
	"core-banking/internal/transaction"

	"github.com/jmoiron/sqlx"
)

func InsertCustomers(tx *sqlx.Tx, data []account.Customer) error {
	query := `
	INSERT INTO customers (
		id, full_name, date_of_birth, nationality,
		email, phone_number, kyc_status, kyc_verified_at,
		risk_level, pep_flag, created_at, updated_at
	)
	VALUES (
		:id, :full_name, :date_of_birth, :nationality,
		:email, :phone_number, :kyc_status, :kyc_verified_at,
		:risk_level, :pep_flag, :created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertAccounts(tx *sqlx.Tx, data []account.Account) error {
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

func InsertTransactions(tx *sqlx.Tx, data []transaction.Transaction) error {
	query := `
	INSERT INTO transactions (
		id, partner_reference_no, reference_no,
		transaction_type, status, amount, currency,
		description, created_at, completed_at
	)
	VALUES (
		:id, :partner_reference_no, :reference_no,
		:transaction_type, :status, :amount, :currency,
		:description, :created_at, :completed_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertTransferDetails(tx *sqlx.Tx, data []transaction.TransferDetail) error {
	query := `
	INSERT INTO transfer_details (
		id, transaction_id, source_account_no, beneficiary_account_no,
		beneficiary_account_name, beneficiary_address, beneficiary_bank_code,
		beneficiary_bank_name, beneficiary_email, customer_reference,
		fee_type, transaction_date, remark, originator_infos, additional_info,
		created_at
	)
	VALUES (
		:id, :transaction_id, :source_account_no, :beneficiary_account_no,
		:beneficiary_account_name, :beneficiary_address, :beneficiary_bank_code,
		:beneficiary_bank_name, :beneficiary_email, :customer_reference,
		:fee_type, :transaction_date, :remark, :originator_infos, :additional_info,
		:created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertJournals(tx *sqlx.Tx, data []transaction.Journal) error {
	query := `
	INSERT INTO journals (
		id, transaction_id, journal_type, status, posted_at, created_at
	)
	VALUES (
		:id, :transaction_id, :journal_type, :status, :posted_at, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertLedgerEntries(tx *sqlx.Tx, data []transaction.LedgerEntry) error {
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

func InsertPayments(tx *sqlx.Tx, data []transaction.Payment) error {
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

func InsertAuditLogs(tx *sqlx.Tx, data []transaction.AuditLog) error {
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

func InsertIdempotencyKeys(tx *sqlx.Tx, data []transaction.IdempotencyKey) error {
	query := `
	INSERT INTO idempotency_keys (
		id, key, request_hash, response_code, response_message, response_body, created_at
	)
	VALUES (
		:id, :key, :request_hash, :response_code, :response_message, :response_body, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertFXRates(tx *sqlx.Tx, data []transaction.FXRate) error {
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
