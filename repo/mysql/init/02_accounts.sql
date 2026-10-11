USE wallet;

CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    created_at DATE NOT NULL,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);
