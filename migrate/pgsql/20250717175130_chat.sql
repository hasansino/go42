-- +goose Up

create table if not exists chat_rooms (
    id bigserial primary key,
    uuid uuid not null,
    name varchar(255) not null,
    type varchar(50) not null default 'public',
    max_users int not null default 100,
    created_by integer not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null,
    constraint fk_chat_rooms_created_by foreign key (created_by) references auth_users(id) on delete cascade
);

create unique index if not exists idx_chat_rooms_uuid on chat_rooms(uuid);
create index if not exists idx_chat_rooms_type on chat_rooms(type);
create index if not exists idx_chat_rooms_created_by on chat_rooms(created_by);
create index if not exists idx_chat_rooms_deleted_at on chat_rooms(deleted_at);

create table if not exists chat_messages (
    id bigserial primary key,
    uuid uuid not null,
    type varchar(50) not null default 'text',
    content text not null,
    user_id integer not null,
    room_id integer not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null,
    constraint fk_chat_messages_user_id foreign key (user_id) references auth_users(id) on delete cascade,
    constraint fk_chat_messages_room_id foreign key (room_id) references chat_rooms(id) on delete cascade
);

create unique index if not exists idx_chat_messages_uuid on chat_messages(uuid);
create index if not exists idx_chat_messages_room_id on chat_messages(room_id);
create index if not exists idx_chat_messages_user_id on chat_messages(user_id);
create index if not exists idx_chat_messages_type on chat_messages(type);
create index if not exists idx_chat_messages_created_at on chat_messages(created_at);
create index if not exists idx_chat_messages_deleted_at on chat_messages(deleted_at);

create table if not exists chat_room_members (
    room_id integer not null,
    user_id integer not null,
    joined_at timestamp not null default current_timestamp,
    left_at timestamp null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    primary key (room_id, user_id),
    constraint fk_chat_room_members_room_id foreign key (room_id) references chat_rooms(id) on delete cascade,
    constraint fk_chat_room_members_user_id foreign key (user_id) references auth_users(id) on delete cascade
);

create index if not exists idx_chat_room_members_user_id on chat_room_members(user_id);
create index if not exists idx_chat_room_members_joined_at on chat_room_members(joined_at);
create index if not exists idx_chat_room_members_left_at on chat_room_members(left_at);

-- +goose Down

drop table if exists chat_room_members;
drop table if exists chat_messages;
drop table if exists chat_rooms;