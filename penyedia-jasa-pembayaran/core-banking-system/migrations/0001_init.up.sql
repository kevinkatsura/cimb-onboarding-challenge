-- Ledger schema for Core Banking System (double-entry bookkeeping)

CREATE TABLE journal_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_ref VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    entry_date      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE journal_lines (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
    account_id       VARCHAR(100) NOT NULL,
    debit            BIGINT NOT NULL DEFAULT 0,
    credit           BIGINT NOT NULL DEFAULT 0,
    currency         VARCHAR(3) NOT NULL DEFAULT 'IDR',
    balance_after    BIGINT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_debit_or_credit CHECK (
        (debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)
    )
);

CREATE TABLE account_ledger_balances (
    account_id      VARCHAR(100) PRIMARY KEY,
    current_balance BIGINT NOT NULL DEFAULT 0,
    currency        VARCHAR(3) NOT NULL DEFAULT 'IDR',
    last_entry_id   UUID REFERENCES journal_entries(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_journal_lines_account ON journal_lines(account_id);
CREATE INDEX idx_journal_lines_entry   ON journal_lines(journal_entry_id);
CREATE INDEX idx_journal_entries_ref   ON journal_entries(transaction_ref);
