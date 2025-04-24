-- +goose Up
create table example_fruits
(
    id         serial
        constraint example_fruits_pk
            primary key,
    created_at timestamptz default now() not null,
    updated_at timestamptz default now() not null,
    deleted_at timestamptz default NULL,
    name       varchar(255)              not null
);


-- +goose Down
drop table example_fruits;