CREATE TABLE IF NOT EXISTS clients (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at DATE NOT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    balance DECIMAL(24,8) NOT NULL DEFAULT 0,
    created_at DATE NOT NULL,
    FOREIGN KEY (client_id) REFERENCES clients(id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(255) PRIMARY KEY,
    account_id_from VARCHAR(255) NOT NULL,
    account_id_to VARCHAR(255) NOT NULL,
    amount DECIMAL(24,8) NOT NULL,
    created_at DATE NOT NULL,
    FOREIGN KEY (account_id_from) REFERENCES accounts(id),
    FOREIGN KEY (account_id_to) REFERENCES accounts(id)
) ENGINE=InnoDB;
