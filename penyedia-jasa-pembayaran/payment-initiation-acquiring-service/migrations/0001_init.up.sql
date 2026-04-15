-- Payment Initiation schema
CREATE SCHEMA IF NOT EXISTS payment_initiation;
SET search_path TO payment_initiation;

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_reference_no VARCHAR(100) NOT NULL,
    reference_no        VARCHAR(100) NOT NULL UNIQUE,
    type                VARCHAR(30) NOT NULL DEFAULT 'intrabank',
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount              BIGINT NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'IDR',
    fee_amount          BIGINT NOT NULL DEFAULT 0,
    fee_type            VARCHAR(10) NOT NULL DEFAULT 'OUR',
    remark              TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transfer_details (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id          UUID NOT NULL REFERENCES transactions(id),
    source_account_no       VARCHAR(20) NOT NULL,
    source_account_name     VARCHAR(255) NOT NULL DEFAULT '',
    beneficiary_account_no  VARCHAR(20) NOT NULL,
    beneficiary_account_name VARCHAR(255) NOT NULL DEFAULT '',
    beneficiary_email       VARCHAR(255) NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotency_keys (
    key         VARCHAR(200) PRIMARY KEY,
    response    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hours')
);

CREATE INDEX idx_transactions_partner_ref ON transactions(partner_reference_no);
CREATE INDEX idx_transactions_ref_no     ON transactions(reference_no);
CREATE INDEX idx_transfer_details_tx     ON transfer_details(transaction_id);
CREATE INDEX idx_idempotency_expires     ON idempotency_keys(expires_at);
