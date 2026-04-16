-- Seed customers and accounts
INSERT INTO customers (id, name, email, phone_no, device_id, device_model, device_os)
VALUES 
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Kevin Katsura', 'kevin@example.com', '+628123456789', 'dev-001', 'iPhone 15', 'iOS 17'),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'Jane Doe', 'jane@example.com', '+628123456780', 'dev-002', 'Pixel 8', 'Android 14');

INSERT INTO accounts (id, customer_id, account_number, product_code, currency, status)
VALUES 
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '80202604160001', 'savings', 'IDR', 'active'),
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '80202604160002', 'savings', 'IDR', 'active');

INSERT INTO account_balances (account_id, available, pending, currency, version)
VALUES 
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 5000000, 0, 'IDR', 1),
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 10000000, 0, 'IDR', 1);
