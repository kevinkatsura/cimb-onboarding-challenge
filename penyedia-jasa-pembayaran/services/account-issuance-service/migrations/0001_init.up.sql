-- Account Issuance schema

CREATE TABLE customers (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(255) NOT NULL DEFAULT '',
    email             VARCHAR(255) NOT NULL DEFAULT '',
    phone_no          VARCHAR(50) NOT NULL DEFAULT '',
    country_code      VARCHAR(5) NOT NULL DEFAULT 'ID',
    device_id         VARCHAR(100) NOT NULL DEFAULT '',
    device_type       VARCHAR(50) NOT NULL DEFAULT '',
    device_model      VARCHAR(100) NOT NULL DEFAULT '',
    device_os         VARCHAR(50) NOT NULL DEFAULT '',
    onboarding_partner VARCHAR(100) NOT NULL DEFAULT '',
    lang              VARCHAR(10) NOT NULL DEFAULT 'en',
    locale            VARCHAR(10) NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     UUID NOT NULL REFERENCES customers(id),
    account_number  VARCHAR(20) NOT NULL UNIQUE,
    product_code    VARCHAR(20) NOT NULL DEFAULT 'savings',
    currency        VARCHAR(3) NOT NULL DEFAULT 'IDR',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE account_balances (
    account_id UUID PRIMARY KEY REFERENCES accounts(id),
    available  BIGINT NOT NULL DEFAULT 0,
    pending    BIGINT NOT NULL DEFAULT 0,
    currency   VARCHAR(3) NOT NULL DEFAULT 'IDR',
    version    INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_customer ON accounts(customer_id);
CREATE INDEX idx_accounts_number   ON accounts(account_number);
CREATE INDEX idx_customers_email   ON customers(email);
