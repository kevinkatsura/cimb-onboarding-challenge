CREATE INDEX IF NOT EXISTS idx_accounts_customer ON accounts(customer_id);                                    -- accounts
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at);                                   -- accounts
CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference_id);                          -- transactions
CREATE INDEX IF NOT EXISTS idx_journal_transaction ON journal_entries(transaction_id)                         -- journal_entries
CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger_entries(account_id);                                  -- ledger_entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_ledger ON ledger_entries(journal_id, account_id, entry_type);    -- ledger_entries