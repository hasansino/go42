-- +goose Up

create table if not exists chat_rooms (
    id integer primary key autoincrement,
    uuid text not null unique,
    name text not null,
    type text not null default 'public',
    max_users integer not null default 100,
    created_by integer not null,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    deleted_at datetime,
    foreign key (created_by) references auth_users(id) on delete cascade
);

create index if not exists idx_chat_rooms_uuid on chat_rooms(uuid);
create index if not exists idx_chat_rooms_type on chat_rooms(type);
create index if not exists idx_chat_rooms_created_by on chat_rooms(created_by);
create index if not exists idx_chat_rooms_deleted_at on chat_rooms(deleted_at);

create table if not exists chat_messages (
    id integer primary key autoincrement,
    uuid text not null unique,
    type text not null default 'text',
    content text not null,
    user_id integer not null,
    room_id integer not null,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    deleted_at datetime,
    foreign key (user_id) references auth_users(id) on delete cascade,
    foreign key (room_id) references chat_rooms(id) on delete cascade
);

create index if not exists idx_chat_messages_uuid on chat_messages(uuid);
create index if not exists idx_chat_messages_room_id on chat_messages(room_id);
create index if not exists idx_chat_messages_user_id on chat_messages(user_id);
create index if not exists idx_chat_messages_type on chat_messages(type);
create index if not exists idx_chat_messages_created_at on chat_messages(created_at);
create index if not exists idx_chat_messages_deleted_at on chat_messages(deleted_at);

create table if not exists chat_room_members (
    room_id integer not null,
    user_id integer not null,
    joined_at datetime not null default current_timestamp,
    left_at datetime,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    primary key (room_id, user_id),
    foreign key (room_id) references chat_rooms(id) on delete cascade,
    foreign key (user_id) references auth_users(id) on delete cascade
);

create index if not exists idx_chat_room_members_user_id on chat_room_members(user_id);
create index if not exists idx_chat_room_members_joined_at on chat_room_members(joined_at);
create index if not exists idx_chat_room_members_left_at on chat_room_members(left_at);

-- +goose Down

drop table if exists chat_room_members;
drop table if exists chat_messages;
drop table if exists chat_rooms;