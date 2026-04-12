CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Customers
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Identity
    full_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    nationality CHAR(2) NOT NULL, -- ISO country code

    -- Contact
    email TEXT UNIQUE NOT NULL, 
    phone_number TEXT UNIQUE NOT NULL,

    -- KYC
    kyc_status VARCHAR(20) NOT NULL CHECK (kyc_status IN ('pending', 'verified', 'rejected')),
    kyc_verified_at TIMESTAMP,

    -- Risk
    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    pep_flag BOOLEAN DEFAULT FALSE, -- politically exposed person

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),

    document_type VARCHAR(30) NOT NULL, -- passport, KTP, etc
    document_number TEXT NOT NULL,
    issuing_country CHAR(2),

    expires_at DATE, 
    created_at TIMESTAMP DEFAULT NOW()
);

-- Accounts 
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),

    account_number VARCHAR(20) UNIQUE NOT NULL,

    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('savings', 'current', 'loan', 'wallet')),

    currency CHAR(3) NOT NULL, -- ISOS currency
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'frozen', 'closed')),

    -- Derived but cached (performance)
    available_balance BIGINT NOT NULL DEFAULT 0,
    pending_balance BIGINT NOT NULL DEFAULT 0,

    overdraft_limit BIGINT DEFAULT 0,

    opened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    deleted_at TIMESTAMP,

    -- Table constraints
    CONSTRAINT check_balance_non_negative CHECK(available_balance >= -overdraft_limit)
);

-- Transactions (logical business event)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 

    partner_reference_no TEXT UNIQUE NOT NULL, -- SNAP PartnerReferenceNo
    reference_no TEXT, -- internal system reference generated after success
    
    transaction_type VARCHAR(30) NOT NULL CHECK (transaction_type IN ('transfer', 'deposit', 'withdrawal', 'payment', 'fee', 'reversal')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),

    amount BIGINT NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL, 

    description TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Transaction Metadata / Transfer Details (SNAP Compliance)
CREATE TABLE IF NOT EXISTS transfer_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID UNIQUE NOT NULL REFERENCES transactions(id),

    source_account_no TEXT NOT NULL,
    beneficiary_account_no TEXT NOT NULL,
    beneficiary_account_name TEXT,
    beneficiary_address TEXT,
    beneficiary_bank_code TEXT,
    beneficiary_bank_name TEXT,
    beneficiary_email TEXT,

    customer_reference TEXT,
    fee_type TEXT CHECK (fee_type IN ('OUR', 'BEN', 'SHA')),
    transaction_date TIMESTAMP,
    remark TEXT,
    
    originator_infos JSONB,
    additional_info JSONB,

    created_at TIMESTAMP DEFAULT NOW()
);

-- Journals (Accounting Intent)
CREATE TABLE IF NOT EXISTS journals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),

    journal_type VARCHAR(30) NOT NULL, -- transfer, fee, settlement
    status VARCHAR(20) NOT NULL DEFAULT 'posted',

    posted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Entries (actual money movement), Ledger lines (double entry core)
CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    journal_id UUID NOT NULL REFERENCES journals(id),
    account_id UUID NOT NULL REFERENCES accounts(id),

    entry_type VARCHAR(10) NOT NULL CHECK (entry_type IN ('debit', 'credit')),

    amount BIGINT NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL,

    -- running balance snapshot (optional optimization)
    balance_after BIGINT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Payments (external movement)
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    transaction_id UUID NOT NULL REFERENCES transactions(id),

    payment_method VARCHAR(30) NOT NULL, -- bank_transfer, card, ewallet
    provider VARCHAR(50), -- VISA, Gopay, etc

    status VARCHAR(20) NOT NULL CHECK (status IN ('initiated', 'processing', 'settled', 'failed')),

    fee_amount BIGINT DEFAULT 0,
    
    metadata JSONB,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Audit logs (compliance)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    actor_id UUID,
    entity_type VARCHAR(50),
    entity_id UUID,

    action VARCHAR(50) NOT NULL, -- create, update, delete

    old_value JSONB,
    new_value JSONB,

    ip_address INET,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    key TEXT UNIQUE NOT NULL,
    request_hash TEXT NOT NULL,

    response_code TEXT,
    response_message TEXT,
    response_body BYTEA,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fx_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    base_currency CHAR(3) NOT NULL,
    quote_currency CHAR(3) NOT NULL,

    rate NUMERIC(20, 8) NOT NULL,
    
    effective_at TIMESTAMP NOT NULL
);

-- Ensure each transaction is balanced
CREATE OR REPLACE FUNCTION validate_transaction_balance(tx_id UUID) 
RETURNS BOOLEAN AS $$
DECLARE 
    total_debit BIGINT;
    total_credit BIGINT;
BEGIN
    SELECT COALESCE(SUM(le.amount), 0)
    INTO total_debit
    FROM ledger_entries le
    JOIN journals j ON le.journal_id = j.id
    WHERE j.transaction_id = tx_id AND le.entry_type = 'debit';

    SELECT COALESCE(SUM(le.amount), 0)
    INTO total_credit
    FROM ledger_entries le
    JOIN journals j ON le.journal_id = j.id
    WHERE j.transaction_id = tx_id AND le.entry_type = 'credit';

    RETURN total_debit = total_credit;
END;
$$ LANGUAGE plpgsql;

-- Derived balance (Materialized View)
CREATE MATERIALIZED VIEW account_balances AS
SELECT
    account_id,
    SUM(
        CASE WHEN entry_type = 'credit' THEN amount
        ELSE -amount
        END
    ) AS balance
FROM ledger_entries
GROUP BY account_id;