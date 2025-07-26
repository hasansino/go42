-- +goose Up

create table if not exists auth_users (
    id integer primary key autoincrement,
    uuid text not null unique,
    password text,
    email text not null unique,
    status text not null default 'active',
    is_system integer not null default 0,
    metadata text,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    deleted_at datetime
);

create index if not exists idx_auth_users_uuid on auth_users(uuid);
create index if not exists idx_auth_users_email on auth_users(email);
create index if not exists idx_auth_users_deleted_at on auth_users(deleted_at);

insert or ignore into auth_users (uuid, password, email, is_system) values
('00000000-0000-0000-0000-000000000000', null, 'admin@system.local', 1);

create table if not exists auth_roles (
    id integer primary key autoincrement,
    name text not null unique,
    description text not null default '',
    is_system integer not null default 0,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    deleted_at datetime
);

create index if not exists idx_auth_roles_deleted_at on auth_roles(deleted_at);

insert or ignore into auth_roles (name, description, is_system) values
('user', 'standard user role with limited access', 0);

create table if not exists auth_permissions (
    id integer primary key autoincrement,
    resource text not null,
    action text not null,
    scope text,
    created_at datetime not null default current_timestamp,
    unique(resource, action, scope)
);

insert or ignore into auth_permissions (resource, action, scope) values
    ('user', 'read_self', null);

create table if not exists auth_role_permissions (
    role_id integer not null,
    permission_id integer not null,
    primary key (role_id, permission_id),
    foreign key (role_id) references auth_roles(id) on delete cascade,
    foreign key (permission_id) references auth_permissions(id) on delete cascade
);

insert or ignore into auth_role_permissions (role_id, permission_id) values
((select id from auth_roles where name = 'user'), (select id from auth_permissions where resource = 'user' and action = 'read_self'));

create index if not exists idx_auth_role_permissions_permission_id on auth_role_permissions(permission_id);

create table if not exists auth_user_roles (
    user_id integer not null,
    role_id integer not null,
    granted_at datetime not null default current_timestamp,
    granted_by integer,
    expires_at datetime,
    primary key (user_id, role_id),
    foreign key (user_id) references auth_users(id) on delete cascade,
    foreign key (role_id) references auth_roles(id) on delete cascade,
    foreign key (granted_by) references auth_users(id) on delete set null
);

create index if not exists idx_auth_user_roles_role_id on auth_user_roles(role_id);
create index if not exists idx_auth_user_roles_expires_at on auth_user_roles(expires_at);

CREATE TABLE IF NOT EXISTS auth_api_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    scopes TEXT DEFAULT NULL,
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    CONSTRAINT fk_api_tokens_user FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_token_lookup ON auth_api_tokens(token, deleted_at, expires_at);

-- +goose Down

drop table if exists auth_api_tokens;
drop table if exists auth_user_roles;
drop table if exists auth_role_permissions;
drop table if exists auth_permissions;
drop table if exists auth_roles;
drop table if exists auth_users;