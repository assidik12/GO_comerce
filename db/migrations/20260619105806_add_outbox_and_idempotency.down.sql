ALTER TABLE transactions
DROP COLUMN idempotency_key;

DROP TABLE IF EXISTS outbox_events;
