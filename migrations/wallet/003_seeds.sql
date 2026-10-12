START TRANSACTION;

INSERT INTO clients (id, name, email, created_at) VALUES
('11111111-1111-4111-8111-111111111111', 'Alice Demo', 'alice@example.test', '2026-10-11'),
('22222222-2222-4222-8222-222222222222', 'Bob Demo', 'bob@example.test', '2026-10-11')
ON DUPLICATE KEY UPDATE id = id;

INSERT INTO accounts (id, client_id, balance, created_at) VALUES
('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 1000.00, '2026-10-11'),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222', 500.00, '2026-10-11')
ON DUPLICATE KEY UPDATE id = id;

COMMIT;
