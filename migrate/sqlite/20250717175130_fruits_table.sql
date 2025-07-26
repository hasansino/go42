-- +goose Up
create table if not exists example_fruits
(
    id         integer primary key autoincrement,
    created_at datetime default (datetime('now')) not null,
    updated_at datetime default (datetime('now')) not null,
    deleted_at datetime default null,
    name       varchar(255) not null unique
);

-- +goose Down
drop table if exists example_fruits;
