-- Indexes
DROP UNIQUE INDEX IF EXISTS idx_unique_ledger;
DROP INDEX IF EXISTS idx_ledger_account;
DROP INDEX IF EXISTS idx_journal_transaction;
DROP INDEX IF EXISTS idx_transactions_reference;
DROP INDEX IF EXISTS idx_accounts_pagination;
DROP INDEX IF EXISTS idx_accounts_deleted_at;
DROP INDEX IF EXISTS idx_accounts_customer;
