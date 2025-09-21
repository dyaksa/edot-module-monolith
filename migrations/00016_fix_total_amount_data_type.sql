-- +goose Up
-- +goose StatementBegin
-- Fix total_amount data type mismatch - change from NUMERIC(18,2) to BIGINT
-- Since we store amounts in cents (smallest currency unit), BIGINT is appropriate
-- Convert existing decimal values to cents by multiplying by 100

-- First, update any existing data to convert to cents
UPDATE orders SET total_amount = total_amount * 100;

-- Then change the column type to BIGINT
ALTER TABLE orders ALTER COLUMN total_amount TYPE BIGINT USING (total_amount::BIGINT);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert back to NUMERIC(18,2) and convert cents back to decimal
ALTER TABLE orders ALTER COLUMN total_amount TYPE NUMERIC(18,2) USING (total_amount::NUMERIC/100);
-- +goose StatementEnd