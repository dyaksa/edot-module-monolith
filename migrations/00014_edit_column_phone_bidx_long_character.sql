-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN phone_bidx TYPE VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN phone_bidx TYPE VARCHAR(50);
-- +goose StatementEnd
