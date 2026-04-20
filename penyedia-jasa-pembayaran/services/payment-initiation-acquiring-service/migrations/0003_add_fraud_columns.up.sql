-- Add fraud detection columns to transactions table
ALTER TABLE payment_initiation.transactions
    ADD COLUMN IF NOT EXISTS fraud_decision VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fraud_event_id VARCHAR(100) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_transactions_fraud_decision
    ON payment_initiation.transactions(fraud_decision);
