-- +goose Up

create table if not exists chat_rooms (
    id bigint unsigned not null auto_increment primary key,
    uuid char(36) not null,
    name varchar(255) not null,
    type varchar(50) not null default 'public',
    max_users int not null default 100,
    created_by bigint unsigned not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp,
    deleted_at timestamp null,
    unique key idx_chat_rooms_uuid (uuid),
    key idx_chat_rooms_type (type),
    key idx_chat_rooms_created_by (created_by),
    key idx_chat_rooms_deleted_at (deleted_at),
    constraint fk_chat_rooms_created_by foreign key (created_by) references auth_users(id) on delete cascade
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

create table if not exists chat_messages (
    id bigint unsigned not null auto_increment primary key,
    uuid char(36) not null,
    type varchar(50) not null default 'text',
    content text not null,
    user_id bigint unsigned not null,
    room_id bigint unsigned not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp,
    deleted_at timestamp null,
    unique key idx_chat_messages_uuid (uuid),
    key idx_chat_messages_room_id (room_id),
    key idx_chat_messages_user_id (user_id),
    key idx_chat_messages_type (type),
    key idx_chat_messages_created_at (created_at),
    key idx_chat_messages_deleted_at (deleted_at),
    constraint fk_chat_messages_user_id foreign key (user_id) references auth_users(id) on delete cascade,
    constraint fk_chat_messages_room_id foreign key (room_id) references chat_rooms(id) on delete cascade
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

create table if not exists chat_room_members (
    room_id bigint unsigned not null,
    user_id bigint unsigned not null,
    joined_at timestamp not null default current_timestamp,
    left_at timestamp null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp,
    primary key (room_id, user_id),
    key idx_chat_room_members_user_id (user_id),
    key idx_chat_room_members_joined_at (joined_at),
    key idx_chat_room_members_left_at (left_at),
    constraint fk_chat_room_members_room_id foreign key (room_id) references chat_rooms(id) on delete cascade,
    constraint fk_chat_room_members_user_id foreign key (user_id) references auth_users(id) on delete cascade
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

-- +goose Down

drop table if exists chat_room_members;
drop table if exists chat_messages;
drop table if exists chat_rooms;