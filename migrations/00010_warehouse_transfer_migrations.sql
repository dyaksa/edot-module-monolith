-- +goose Up
-- +goose StatementBegin
CREATE TYPE transfer_status AS ENUM ('REQUESTED', 'APPROVED', 'IN_TRANSIT', 'COMPLETED', 'CANCELLED');
CREATE TABLE warehouse_transfers (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    to_warehouse_id   UUID NOT NULL REFERENCES warehouses(id),
    status            transfer_status NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE warehouse_transfers;
DROP TYPE transfer_status;
-- +goose StatementEnd
