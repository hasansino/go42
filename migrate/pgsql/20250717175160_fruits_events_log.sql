-- +goose Up
create table if not exists example_fruits_events_log
(
    id          text primary key,
    occurred_at timestamp not null,
    created_at  timestamp default now() not null,
    fruit_id    integer not null,
    event_type  varchar(255) not null,
    data        varchar(255) not null,
    metadata    varchar(1000) not null
);

-- +goose Down
drop table if exists example_fruits_events_log;