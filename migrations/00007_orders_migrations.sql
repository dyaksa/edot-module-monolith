-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM ('PENDING', 'AWAITING_PAYMENT', 'PAID', 'CANCELLED', 'EXPIRED', 'FULFILLED');
CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    shop_id         UUID NOT NULL REFERENCES shops(id),
    status          order_status NOT NULL,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    reservation_expires_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
DROP TYPE order_status;
-- +goose StatementEnd
