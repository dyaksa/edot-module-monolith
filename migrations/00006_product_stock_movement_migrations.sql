-- +goose Up
-- +goose StatementBegin
CREATE TYPE movement_type AS ENUM ('IN', 'OUT', 'RESERVE', 'RELEASE', 'TRANSFER_IN', 'TRANSFER_OUT');
CREATE TABLE stock_movements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id    UUID NOT NULL REFERENCES products(id),
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id),
    type          movement_type NOT NULL,
    qty           INT NOT NULL,
    ref_type      VARCHAR(50),
    ref_id        UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE stock_movements;
DROP TYPE movement_type;
-- +goose StatementEnd
