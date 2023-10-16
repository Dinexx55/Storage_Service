-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS store_versions (
version_id SERIAL PRIMARY KEY,
store_id INT NOT NULL,
version_number INT NOT NULL,
creator_login VARCHAR(255) NOT NULL,
owner_name VARCHAR(255) NOT NULL,
opening_time TIME NOT NULL,
closing_time TIME NOT NULL,
created_at VARCHAR(255) NOT NULL,
is_last bool NOT NULL,
FOREIGN KEY (store_id) REFERENCES stores (store_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS store_versions;
-- +goose StatementEnd