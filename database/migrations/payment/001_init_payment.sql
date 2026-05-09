CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(32) NOT NULL UNIQUE,
    customer_pan VARCHAR(32) NOT NULL,
    amount NUMERIC(18, 2) NOT NULL,
    transaction_datetime TIMESTAMPTZ NOT NULL,
    rrn VARCHAR(32) NOT NULL,
    bill_number VARCHAR(64) NOT NULL UNIQUE,
    customer_name VARCHAR(255),
    merchant_id VARCHAR(32) NOT NULL,
    merchant_name VARCHAR(255) NOT NULL,
    merchant_city VARCHAR(255) NOT NULL,
    currency_code VARCHAR(8) NOT NULL,
    payment_status VARCHAR(8) NOT NULL,
    payment_description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);