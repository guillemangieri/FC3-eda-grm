-- Upgrade databases created by the original INT schema, preserving existing values.
ALTER TABLE accounts MODIFY balance DECIMAL(24,8) NOT NULL DEFAULT 0;
ALTER TABLE transactions MODIFY amount DECIMAL(24,8) NOT NULL;
