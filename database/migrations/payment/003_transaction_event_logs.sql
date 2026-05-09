CREATE TABLE IF NOT EXISTS transaction_event_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(32) NOT NULL,
    merchant_id VARCHAR(32) NOT NULL,
    bill_number VARCHAR(64) NOT NULL,
    status VARCHAR(8) NOT NULL,
    event_payload JSONB NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transaction_event_logs_request_id ON transaction_event_logs (request_id);

CREATE INDEX IF NOT EXISTS idx_transaction_event_logs_merchant_id ON transaction_event_logs (merchant_id);

CREATE INDEX IF NOT EXISTS idx_transaction_event_logs_bill_number ON transaction_event_logs (bill_number);

CREATE INDEX IF NOT EXISTS idx_transaction_event_logs_consumed_at ON transaction_event_logs (consumed_at DESC);