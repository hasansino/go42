-- +goose Up
create table transactional_outbox (
    id char(36) primary key,
    aggregate_id int not null,
    aggregate_type varchar(100) not null,
    topic varchar(255) not null,
    payload text not null,
    created_at timestamp default current_timestamp,
    processed_at timestamp null,
    status enum('pending', 'processed', 'failed') not null,
    retry_count int not null,
    max_retries int not null,
    last_error text not null ,
    metadata text not null
);

create index transactional_outbox_publisher ON transactional_outbox (status);

-- +goose Down
DROP TABLE transactional_outbox;