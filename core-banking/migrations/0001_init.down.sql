-- Derived balance (Materialized View)
DROP MATERIALIZED VIEW IF EXISTS account_balances;

-- Ensure each transaction is balanced
DROP FUNCTION IF EXISTS validate_transaction_balance(tx_id UUID);

-- Indexes
DROP INDEX IF EXISTS idx_journal_entries_transaction_id;
DROP INDEX IF EXISTS idx_ledger_entries_journal_id;
DROP INDEX IF EXISTS idx_ledger_entries_account_id;
DROP INDEX IF EXISTS idx_accounts_customer_id;

-- Tables
DROP TABLE IF EXISTS fx_rates;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS journal_entries;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS customer_documents;
DROP TABLE IF EXISTS customers;

-- Extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pgcrypto";