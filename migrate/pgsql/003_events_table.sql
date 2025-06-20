-- +goose Up
create table example_fruits_events
(
    id         serial
        constraint example_fruits_events_pk
            primary key,
    created_at timestamp default now() not null,
    data      varchar(255)             not null
);

-- +goose Down
drop table example_fruits_events;
