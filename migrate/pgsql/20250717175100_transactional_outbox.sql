-- +goose Up
create table if not exists transactional_outbox (
    id uuid primary key,
    aggregate_id integer not null,
    aggregate_type varchar(100) not null,
    topic varchar(255) not null,
    payload text not null,
    created_at timestamp with time zone default current_timestamp,
    processed_at timestamp with time zone null,
    status varchar(20) not null check (status in ('pending', 'processed', 'failed')),
    retry_count integer not null,
    max_retries integer not null,
    last_error text not null,
    metadata text not null
);

create index if not exists transactional_outbox_publisher ON transactional_outbox (status);

-- +goose Down
drop table if exists transactional_outbox;
