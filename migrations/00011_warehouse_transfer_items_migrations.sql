-- +goose Up
-- +goose StatementBegin
CREATE TABLE warehouse_transfer_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES warehouse_transfers(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id),
    qty         INT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE warehouse_transfer_items;
-- +goose StatementEnd
