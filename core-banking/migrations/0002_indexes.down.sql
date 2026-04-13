-- idempotency_keys
DROP INDEX IF EXISTS idx_idempotency_key;

-- account_transactions
DROP INDEX IF EXISTS idx_acc_tx_cursor;
DROP INDEX IF EXISTS idx_acc_tx_transaction;
DROP INDEX IF EXISTS idx_acc_tx_account;

-- ledger_entries
DROP INDEX IF EXISTS idx_ledger_account_cursor;
DROP INDEX IF EXISTS idx_ledger_account;
DROP INDEX IF EXISTS idx_ledger_transaction;

-- transfer_details
DROP INDEX IF EXISTS idx_transfer_details_transaction;

-- transactions
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_partner_ref;
DROP INDEX IF EXISTS idx_transactions_reference;

-- accounts
DROP INDEX IF EXISTS idx_accounts_pagination;
DROP INDEX IF EXISTS idx_accounts_number;
DROP INDEX IF EXISTS idx_accounts_customer;

-- customers
DROP INDEX IF EXISTS idx_customers_ext_id;
DROP INDEX IF EXISTS idx_customers_partner_ref;
DROP INDEX IF EXISTS idx_customers_phone;
DROP INDEX IF EXISTS idx_customers_email;