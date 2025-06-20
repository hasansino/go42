-- +goose Up
create table example_fruits
(
    id         serial
        constraint example_fruits_pk
            primary key,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    deleted_at timestamp default NULL,
    name       varchar(255)              not null
);

create unique index example_fruits_name_unique ON example_fruits (name);

-- +goose Down
drop table example_fruits;
