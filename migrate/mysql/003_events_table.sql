-- +goose Up
CREATE TABLE example_fruits_events
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    data       VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE example_fruits_events;