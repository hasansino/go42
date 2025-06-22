-- +goose Up
CREATE TABLE example_fruits
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name       VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX example_fruits_name_unique ON example_fruits (name);

-- +goose Down
DROP TABLE example_fruits;