-- +goose Up
create table example_fruits_events_log
(
    id          varchar(255) primary key,
    occurred_at datetime not null,
    created_at  datetime default current_timestamp not null,
    fruit_id    integer not null,
    event_type  varchar(255) not null,
    data        varchar(255) not null,
    metadata    varchar(1000) not null
);

-- +goose Down
drop table example_fruits_events_log;