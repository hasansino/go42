-- +goose Up
create table if not exists transactional_outbox (
    id char(36) primary key,
    aggregate_id int not null,
    aggregate_type varchar(100) not null,
    topic varchar(255) not null,
    payload text null,
    created_at timestamp default current_timestamp,
    processed_at timestamp null,
    status enum('pending', 'processed', 'failed') not null,
    retry_count int not null,
    max_retries int not null,
    last_error text not null ,
    metadata text null,
    key transactional_outbox_publisher (status)
);

-- +goose Down
drop table if exists transactional_outbox;
