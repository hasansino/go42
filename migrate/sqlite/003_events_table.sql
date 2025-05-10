-- +goose Up
create table example_events
(
    id         integer primary key autoincrement,
    created_at datetime default (datetime('now')) not null,
    data       varchar(255) not null
);

-- +goose Down
drop table example_events;