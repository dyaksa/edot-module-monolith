-- +goose Up
-- +goose StatementBegin
CREATE TYPE reservation_status AS ENUM ('PENDING', 'COMMITTED', 'RELEASED', 'EXPIRED');
CREATE TABLE stock_reservations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    qty         INT NOT NULL,
    status      reservation_status NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reservations_status_expires
    ON stock_reservations(status, expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stock_reservations;
DROP TYPE reservation_status;
-- +goose StatementEnd
