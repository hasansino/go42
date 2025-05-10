-- +goose Up
create table example_events
(
    id         serial
        constraint example_events_pk
            primary key,
    created_at timestamptz default now() not null,
    data      varchar(255)               not null
);

-- +goose Down
drop table example_events;