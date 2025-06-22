-- +goose Up
create table transactional_outbox (
    id text primary key,
    aggregate_id integer not null,
    aggregate_type text not null,
    topic text not null,
    payload text not null,
    created_at datetime default current_timestamp,
    processed_at datetime null,
    status text not null check(status in ('pending', 'processed', 'failed')),
    retry_count integer not null,
    max_retries integer not null,
    last_error text not null,
    metadata text not null
);

create unique index transactional_outbox_publisher ON transactional_outbox (status);

-- +goose Down
drop table transactional_outbox;
