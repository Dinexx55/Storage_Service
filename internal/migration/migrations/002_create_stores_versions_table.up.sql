CREATE TABLE store_versions (
                               version_id SERIAL PRIMARY KEY,
                               store_id INT NOT NULL,
                               name VARCHAR(255) NOT NULL,
                               address VARCHAR(255) NOT NULL,
                               owner_name VARCHAR(255) NOT NULL,
                               opening_time TIME NOT NULL,
                               closing_time TIME NOT NULL,
                               created_at TIMESTAMPTZ DEFAULT NOW(),
                               FOREIGN KEY (store_id) REFERENCES stores (store_id)
);