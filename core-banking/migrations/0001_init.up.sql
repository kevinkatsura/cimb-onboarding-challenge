CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    full_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    nationality CHAR(2) NOT NULL,

    email TEXT UNIQUE NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,

    -- SNAP / Compliance Identity
    partner_reference_no TEXT, -- Transaction identifier on service consumer system
    country_code CHAR(2),      -- Requestor’s country code (ISO 3166-1 alpha-2)
    external_customer_id TEXT, -- account ID of the customer on consumer system
    
    -- Device Context
    device_os TEXT,           -- Device’s OS
    device_os_version TEXT,   -- Device’s OS version
    device_model TEXT,        -- Device’s model
    device_manufacturer TEXT, -- Device’s manufacturer
    
    -- Localization & Onboarding
    lang VARCHAR(8),          -- language support parameter
    locale VARCHAR(5),        -- Locale and language selected in app
    onboarding_partner TEXT,  -- Onboarding partner of the customer
    redirect_url TEXT,        -- Merchant call back URL
    
    -- Auth & Flow
    scopes TEXT,              -- The scopes of the authorization
    seamless_data TEXT,       -- structure for mobile/verification info (URLencoded)
    seamless_sign TEXT,       -- signature data for seamlessData (URLencoded)
    state VARCHAR(32),        -- state
    
    -- Merchant Identity
    merchant_id TEXT,         -- Merchant identifier (unique)
    sub_merchant_id TEXT,     -- Sub merchant ID
    terminal_type TEXT,       -- terminal type / redirect source
    
    -- Extensibility
    additional_info JSONB,    -- Additional info for custom use

    kyc_status VARCHAR(20) NOT NULL CHECK (kyc_status IN ('pending','verified','rejected')),
    kyc_verified_at TIMESTAMP,

    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    pep_flag BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    code VARCHAR(50) PRIMARY KEY,

    name TEXT NOT NULL,
    currency CHAR(3) NOT NULL,

    min_balance BIGINT DEFAULT 0,
    overdraft_limit BIGINT DEFAULT 0,
    daily_limit BIGINT,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    customer_id UUID NOT NULL REFERENCES customers(id),

    account_number VARCHAR(34) UNIQUE NOT NULL,

    product_code VARCHAR(50) NOT NULL REFERENCES products(code),

    currency CHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending','active','frozen','closed')),

    opened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    partner_reference_no TEXT UNIQUE NOT NULL,
    reference_no TEXT,

    transaction_type VARCHAR(30) NOT NULL 
        CHECK (transaction_type IN ('transfer','deposit','withdrawal','payment','fee','reversal')),

    status VARCHAR(20) NOT NULL 
        CHECK (status IN ('initiated','completed','failed','reversed')),

    amount BIGINT NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE TABLE transfer_details (
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

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    transaction_id UUID NOT NULL REFERENCES transactions(id),
    account_id UUID NOT NULL REFERENCES accounts(id),

    entry_type VARCHAR(10) NOT NULL 
        CHECK (entry_type IN ('debit','credit')),

    amount BIGINT NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE account_balances (
    account_id UUID PRIMARY KEY REFERENCES accounts(id),

    available_balance BIGINT NOT NULL,
    pending_balance BIGINT NOT NULL DEFAULT 0,

    last_updated TIMESTAMP NOT NULL
);

CREATE TABLE account_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    account_id UUID NOT NULL REFERENCES accounts(id),
    transaction_id UUID NOT NULL REFERENCES transactions(id),

    direction VARCHAR(10) CHECK (direction IN ('in','out')),
    amount BIGINT NOT NULL,

    created_at TIMESTAMP NOT NULL
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES customers(id), -- User who performed the action
    entity_type VARCHAR(50) NOT NULL,       -- e.g., 'account', 'transaction'
    entity_id UUID,                         -- ID of the affected entity
    action VARCHAR(50) NOT NULL,             -- e.g., 'create', 'update', 'delete', 'transfer'
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT UNIQUE NOT NULL,
    response_code VARCHAR(10),
    response_message TEXT,
    response_body JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE fx_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency CHAR(3) NOT NULL,
    quote_currency CHAR(3) NOT NULL,
    rate NUMERIC(18, 6) NOT NULL,
    effective_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE (base_currency, quote_currency, effective_at)
);