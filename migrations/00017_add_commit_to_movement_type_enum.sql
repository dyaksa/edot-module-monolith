-- +goose Up
-- +goose StatementBegin
-- Add COMMIT to movement_type enum for payment confirmation stock movements
ALTER TYPE movement_type ADD VALUE 'COMMIT';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Note: PostgreSQL doesn't support removing enum values directly
-- This would require recreating the enum and updating all references
-- For development, you can drop and recreate, but be careful in production
-- +goose StatementEnd