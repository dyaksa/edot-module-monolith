-- +goose Up
-- +goose StatementBegin
CREATE TABLE product_stock (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    on_hand       INT NOT NULL DEFAULT 0,
    reserved      INT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(product_id, warehouse_id),
    CHECK (on_hand >= reserved AND on_hand >= 0 AND reserved >= 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE product_stock;
-- +goose StatementEnd
