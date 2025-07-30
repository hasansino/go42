-- +goose Up
create table if not exists transactional_outbox (
    id text primary key,
    aggregate_id integer not null,
    aggregate_type text not null,
    topic text not null,
    payload text null,
    created_at datetime default current_timestamp,
    processed_at datetime null,
    status text not null check(status in ('pending', 'processed', 'failed')),
    retry_count integer not null,
    max_retries integer not null,
    last_error text not null,
    metadata text null
);

create index if not exists transactional_outbox_publisher ON transactional_outbox (status);

-- +goose Down
drop table if exists transactional_outbox;
