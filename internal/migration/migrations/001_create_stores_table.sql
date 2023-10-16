-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS stores (
store_id SERIAL PRIMARY KEY,
name VARCHAR(255) NOT NULL,
address VARCHAR(255) NOT NULL,
creator_login VARCHAR(255) NOT NULL,
owner_name VARCHAR(255) NOT NULL,
opening_time TIME NOT NULL,
closing_time TIME NOT NULL,
created_at VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stores;
-- +goose StatementEnd

