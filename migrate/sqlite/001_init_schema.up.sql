-- +goose Up
create table example_fruits
(
    id         integer primary key autoincrement,
    created_at datetime default (datetime('now')) not null,
    updated_at datetime default (datetime('now')) not null,
    deleted_at datetime default null,
    name       varchar(255) not null
);

-- +goose Down
drop table example_fruits;