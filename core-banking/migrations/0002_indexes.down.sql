-- Reverse Order
DROP INDEX IF EXISTS idx_transfer_details_transaction;
DROP INDEX IF EXISTS idx_unique_ledger;
DROP INDEX IF EXISTS idx_ledger_account_cursor;
DROP INDEX IF EXISTS idx_ledger_journal;
DROP INDEX IF EXISTS idx_ledger_account;
DROP INDEX IF EXISTS idx_journals_transaction;
DROP INDEX IF EXISTS idx_transactions_partner_ref;
DROP INDEX IF EXISTS idx_transactions_reference;
DROP INDEX IF EXISTS idx_accounts_pagination;
DROP INDEX IF EXISTS idx_accounts_deleted_at;
DROP INDEX IF EXISTS idx_accounts_customer;