-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ALTER COLUMN phone TYPE BYTEA USING phone::BYTEA,
    ALTER COLUMN email TYPE BYTEA USING email::BYTEA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    ALTER COLUMN phone TYPE TEXT,
    ALTER COLUMN email TYPE TEXT;
-- +goose StatementEnd
