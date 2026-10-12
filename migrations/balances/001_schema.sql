CREATE TABLE IF NOT EXISTS account_balances (
    account_id VARCHAR(255) PRIMARY KEY,
    balance DECIMAL(24,8) NOT NULL,
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS consumer_offsets (
    topic VARCHAR(255) NOT NULL,
    partition_id INT NOT NULL,
    last_offset BIGINT NOT NULL,
    PRIMARY KEY (topic, partition_id)
) ENGINE=InnoDB;
