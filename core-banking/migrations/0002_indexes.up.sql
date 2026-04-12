-- accounts
CREATE INDEX IF NOT EXISTS idx_accounts_customer ON accounts(customer_id);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts(deleted_at);
CREATE INDEX IF NOT EXISTS idx_accounts_pagination ON accounts (created_at DESC, id DESC) WHERE deleted_at IS NULL;

-- transactions
CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference_no);
CREATE INDEX IF NOT EXISTS idx_transactions_partner_ref ON transactions(partner_reference_no);

-- journals
CREATE INDEX IF NOT EXISTS idx_journals_transaction ON journals(transaction_id);

-- ledger_entries
CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_ledger_journal ON ledger_entries(journal_id);
CREATE INDEX IF NOT EXISTS idx_ledger_account_cursor ON ledger_entries(account_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_unique_ledger ON ledger_entries(journal_id, account_id, entry_type);

-- transfer_details
CREATE INDEX IF NOT EXISTS idx_transfer_details_transaction ON transfer_details(transaction_id);