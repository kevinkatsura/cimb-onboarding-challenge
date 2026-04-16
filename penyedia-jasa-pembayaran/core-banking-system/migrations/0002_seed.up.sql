-- Initialize ledger balances for seeded accounts
INSERT INTO account_ledger_balances (account_id, current_balance, currency, updated_at)
VALUES 
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 5000000, 'IDR', NOW()),
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 10000000, 'IDR', NOW()),
('FEE_REVENUE', 0, 'IDR', NOW()),
('SYSTEM_EQUITY', -15000000, 'IDR', NOW());

-- Create journal entries for the initial balances
INSERT INTO journal_entries (id, transaction_ref, description)
VALUES 
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'INIT-001', 'Initial balance for Kevin'),
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'INIT-002', 'Initial balance for Jane');

INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, balance_after)
VALUES 
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 5000000, 0, 5000000),
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'SYSTEM_EQUITY', 0, 5000000, -5000000),
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 10000000, 0, 10000000),
('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'SYSTEM_EQUITY', 0, 10000000, -15000000);
