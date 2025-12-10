-- +goose Up
ALTER TABLE users ADD COLUMN session_token VARCHAR(255);

-- +goose Down
ALTER TABLE users DROP COLUMN session_token;
