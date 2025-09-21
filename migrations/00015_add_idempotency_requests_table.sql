-- +goose Up
-- +goose StatementBegin
CREATE TABLE idempotency_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(128) NOT NULL, -- Idempotency-Key dari header
    endpoint        VARCHAR(128) NOT NULL, -- misal: POST:/orders/checkout
    payload_hash    TEXT NOT NULL,         -- hash dari body request (SHA256)
    order_id        UUID,                  -- reference hasil order
    response_body   JSONB,                 -- optional: cache response siap kirim ulang
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(key, endpoint)
);
CREATE INDEX idx_idempotency_requests_key_endpoint ON idempotency_requests(key, endpoint);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS idempotency_requests;
-- +goose StatementEnd
