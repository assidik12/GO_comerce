CREATE TABLE outbox_events (
    id VARCHAR(36) PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    topic VARCHAR(255) NOT NULL,
    payload JSON NOT NULL,
    status ENUM('PENDING', 'PROCESSED') NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status)
);

ALTER TABLE transactions
ADD COLUMN idempotency_key VARCHAR(255) UNIQUE DEFAULT NULL;
