START TRANSACTION;
INSERT INTO account_balances (account_id, balance) VALUES
('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 1000.00),
('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 500.00)
ON DUPLICATE KEY UPDATE account_id = account_id;
COMMIT;
