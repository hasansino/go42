-- +goose Up
create table if not exists example_fruits
(
    id         serial primary key,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    deleted_at timestamp default NULL,
    name       varchar(255)              not null
);

create unique index if not exists example_fruits_name_unique on example_fruits (name);

-- +goose Down
drop table if exists example_fruits;
