CREATE TABLE IF NOT EXISTS equity_tracking_tb (
    id                SERIAL PRIMARY KEY,
    broker_account_id BIGINT         NOT NULL REFERENCES broker_accounts_tb (id),
    equity            NUMERIC(19, 4) NOT NULL,
    created_at        TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

