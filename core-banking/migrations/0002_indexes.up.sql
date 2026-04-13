-- customers
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone_number);
CREATE INDEX IF NOT EXISTS idx_customers_partner_ref ON customers(partner_reference_no);
CREATE INDEX IF NOT EXISTS idx_customers_ext_id ON customers(external_customer_id);

-- accounts
CREATE INDEX IF NOT EXISTS idx_accounts_customer ON accounts(customer_id);
CREATE INDEX IF NOT EXISTS idx_accounts_number ON accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_accounts_pagination ON accounts (created_at DESC, id DESC) WHERE status != 'closed';

-- transactions
CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference_no);
CREATE INDEX IF NOT EXISTS idx_transactions_partner_ref ON transactions(partner_reference_no);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);

-- transfer_details
CREATE INDEX IF NOT EXISTS idx_transfer_details_transaction ON transfer_details(transaction_id);

-- ledger_entries
CREATE INDEX IF NOT EXISTS idx_ledger_transaction ON ledger_entries(transaction_id);
CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_ledger_account_cursor ON ledger_entries(account_id, created_at DESC, id DESC);

-- account_transactions (History)
CREATE INDEX IF NOT EXISTS idx_acc_tx_account ON account_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_acc_tx_transaction ON account_transactions(transaction_id);
CREATE INDEX IF NOT EXISTS idx_acc_tx_cursor ON account_transactions(account_id, created_at DESC);

-- idempotency_keys
CREATE INDEX IF NOT EXISTS idx_idempotency_key ON idempotency_keys(key);