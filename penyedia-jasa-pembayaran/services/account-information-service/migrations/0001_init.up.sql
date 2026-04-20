CREATE TABLE IF NOT EXISTS accounts (
    account_number VARCHAR(50) PRIMARY KEY,
    account_id VARCHAR(100) UNIQUE NOT NULL,
    customer_id VARCHAR(100) NOT NULL,
    product_code VARCHAR(50),
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    transaction_ref VARCHAR(100) UNIQUE NOT NULL,
    source_account_number VARCHAR(50),
    beneficiary_account_number VARCHAR(50),
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_source ON transactions(source_account_number);
CREATE INDEX idx_transactions_beneficiary ON transactions(beneficiary_account_number);
CREATE INDEX idx_accounts_id ON accounts(account_id);
