-- INDEX
DROP INDEX IF EXISTS idx_entries_account_id;
DROP INDEX IF EXISTS idx_entries_transaction_id;
DROP INDEX IF EXISTS idx_accounts_customer_id;

-- MATERIALIZED VIEW
DROP MATERIALIZED VIEW IF EXISTS account_balances;
DROP FUNCTION IF EXISTS validate_transaction_balance(tx_id UUID);

-- TABLE
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS entries
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS customers;