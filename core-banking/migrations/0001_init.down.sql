-- Reverse Order
DROP MATERIALIZED VIEW IF EXISTS account_balances;
DROP FUNCTION IF EXISTS validate_transaction_balance(UUID);

-- Drop Tables
DROP TABLE IF EXISTS fx_rates;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS journals;
DROP TABLE IF EXISTS transfer_details;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS customer_documents;
DROP TABLE IF EXISTS customers;

-- Extensions are usually safe to keep, but can be dropped if needed
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";