package seeder

import (
	"core-banking/internal/account"
	"core-banking/internal/transaction"

	"github.com/jmoiron/sqlx"
)

func InsertProducts(tx *sqlx.Tx, data []account.Product) error {
	query := `
	INSERT INTO products (
		code, name, currency, min_balance, overdraft_limit, daily_limit, created_at
	)
	VALUES (
		:code, :name, :currency, :min_balance, :overdraft_limit, :daily_limit, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertCustomers(tx *sqlx.Tx, data []account.Customer) error {
	query := `
	INSERT INTO customers (
		id, full_name, date_of_birth, nationality,
		email, phone_number, kyc_status, kyc_verified_at,
		risk_level, pep_flag, 
		partner_reference_no, country_code, external_customer_id,
		device_os, device_os_version, device_model, device_manufacturer,
		lang, locale, onboarding_partner, redirect_url,
		scopes, seamless_data, seamless_sign, state,
		merchant_id, sub_merchant_id, terminal_type,
		additional_info,
		created_at, updated_at
	)
	VALUES (
		:id, :full_name, :date_of_birth, :nationality,
		:email, :phone_number, :kyc_status, :kyc_verified_at,
		:risk_level, :pep_flag,
		:partner_reference_no, :country_code, :external_customer_id,
		:device_os, :device_os_version, :device_model, :device_manufacturer,
		:lang, :locale, :onboarding_partner, :redirect_url,
		:scopes, :seamless_data, :seamless_sign, :state,
		:merchant_id, :sub_merchant_id, :terminal_type,
		:additional_info,
		:created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertAccounts(tx *sqlx.Tx, data []account.Account) error {
	query := `
	INSERT INTO accounts (
		id, customer_id, account_number, product_code,
		currency, status, opened_at, created_at, updated_at
	)
	VALUES (
		:id, :customer_id, :account_number, :product_code,
		:currency, :status, :opened_at, :created_at, :updated_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertTransactions(tx *sqlx.Tx, data []transaction.Transaction) error {
	query := `
	INSERT INTO transactions (
		id, partner_reference_no, reference_no,
		transaction_type, status, amount, currency,
		created_at, completed_at
	)
	VALUES (
		:id, :partner_reference_no, :reference_no,
		:transaction_type, :status, :amount, :currency,
		:created_at, :completed_at
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

func InsertLedgerEntries(tx *sqlx.Tx, data []transaction.LedgerEntry) error {
	query := `
	INSERT INTO ledger_entries (
		id, transaction_id, account_id,
		entry_type, amount, currency,
		created_at
	)
	VALUES (
		:id, :transaction_id, :account_id,
		:entry_type, :amount, :currency,
		:created_at
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
		id, key, response_code, response_message, response_body, created_at
	)
	VALUES (
		:id, :key, :response_code, :response_message, :response_body, :created_at
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
	if err != nil {
		return err
	}
	return nil
}

func InsertAccountBalances(tx *sqlx.Tx, data []account.AccountBalance) error {
	query := `
	INSERT INTO account_balances (
		account_id, available_balance, pending_balance, last_updated
	)
	VALUES (
		:account_id, :available_balance, :pending_balance, :last_updated
	)`
	_, err := tx.NamedExec(query, data)
	return err
}

func InsertAccountTransactions(tx *sqlx.Tx, data []account.AccountTransaction) error {
	query := `
	INSERT INTO account_transactions (
		id, account_id, transaction_id, direction, amount, created_at
	)
	VALUES (
		:id, :account_id, :transaction_id, :direction, :amount, :created_at
	)`
	_, err := tx.NamedExec(query, data)
	return err
}
