-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN email_bidx VARCHAR(255) UNIQUE,
ADD COLUMN phone_bidx VARCHAR(50) UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN email_bidx,
DROP COLUMN phone_bidx;
-- +goose StatementEnd
