CREATE TABLE store_versions (
                               version_id SERIAL PRIMARY KEY,
                               store_id INT NOT NULL,
                               creator_login VARCHAR(255) NOT NULL,
                               opening_time TIME NOT NULL,
                               store_owner_name VARCHAR(255) NOT NULL,
                               opening_time TIME NOT NULL,
                               closing_time TIME NOT NULL,
                               created_at VARCHAR(255) NOT NULL,
                               FOREIGN KEY (store_id) REFERENCES stores (store_id)
);