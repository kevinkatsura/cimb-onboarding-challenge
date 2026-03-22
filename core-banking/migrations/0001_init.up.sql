CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Customers
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Accounts 
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID REFERENCES customers(id),
    balance BIGINT NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'suspend', 'closed'))
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_accounts_customer_id ON accounts(customer_id);

-- Transactions (logical event)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY, 
    reference_id TEXT UNIQUE NOT NULL, -- idempotency key
    type TEXT NOT NULL, -- transfer, deposit, withdrawal
    status TEXT NOT NULL CHECK (status IN ('pending', 'posted', 'failed'))
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Entries (actual money movement)
CREATE TABLE IF NOT EXISTS entries (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    direction TEXT NOT NULL CHECK (direction IN ('debit', 'credit')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_entries_transaction_id ON entries(transaction_id);
CREATE INDEX idx_entries_account_id ON entries(account_id);


-- Ensure each transaction is balanced
CREATE FUNCTION validate_transaction_balance(tx_id UUID) 
RETURNS BOOLEAN AS $$
DECLARE 
    total_debit BIGINT;
    total_credit BIGINT;
BEGIN
    SELECT COALESCE(SUM(amount), 0)
    INTO total_debit
    FROM entries
    WHERE transaction_id = tx_id AND direction = 'debit';

    SELECT COALESCE(SUM(amount), 0)
    INTO total_credit
    FROM entries
    WHERE transaction_id = tx_id AND direction = 'credit';

    RETURN total_debit = total_credit;
END;
$$ LANGUAGE plpgsql;

-- Derived balance (Materialized View)
CREATE MATERIALIZED VIEW account_balances AS
SELECT
    account_id,
    SUM(
        CASE WHEN direction = 'credit' THEN amount
        ELSE -amount
        END
    ) AS balance
FROM entries
GROUP BY account_id;

-- Audit Logging
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    entity_type TEXT,
    entity_id UUID,
    action TEXT, 
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
)