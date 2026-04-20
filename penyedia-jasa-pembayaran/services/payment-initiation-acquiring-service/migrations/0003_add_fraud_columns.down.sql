-- Remove fraud detection columns from transactions table
DROP INDEX IF EXISTS payment_initiation.idx_transactions_fraud_decision;

ALTER TABLE payment_initiation.transactions
    DROP COLUMN IF EXISTS fraud_decision,
    DROP COLUMN IF EXISTS fraud_event_id;
