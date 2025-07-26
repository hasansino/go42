-- +goose Up
create table if not exists example_fruits
(
    id         int auto_increment primary key,
    created_at timestamp default current_timestamp not null,
    updated_at timestamp default current_timestamp not null,
    deleted_at timestamp null default null,
    name       varchar(255) not null,
    unique key example_fruits_name_unique (name)
);

-- +goose Down
drop table if exists example_fruits;