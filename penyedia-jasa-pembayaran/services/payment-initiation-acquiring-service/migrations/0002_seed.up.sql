-- Seed a few historical transactions
INSERT INTO transactions (id, partner_reference_no, reference_no, type, status, amount, currency, fee_amount, fee_type, remark)
VALUES 
('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'p-ref-001', 'REF001', 'intrabank', 'completed', 100000, 'IDR', 2500, 'OUR', 'Lunch money'),
('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'p-ref-002', 'REF002', 'intrabank', 'completed', 50000, 'IDR', 2500, 'OUR', 'Coffee');

INSERT INTO transfer_details (id, transaction_id, source_account_no, source_account_name, beneficiary_account_no, beneficiary_account_name)
VALUES 
('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '80202604160002', 'Jane Doe', '80202604160001', 'Kevin Katsura'),
('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '80202604160001', 'Kevin Katsura', '80202604160002', 'Jane Doe');
