-- Reverse Order (Dependencies first)
DROP TABLE IF EXISTS fx_rates;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS account_transactions;
DROP TABLE IF EXISTS account_balances;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS transfer_details;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS customers;

-- Drop Extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";