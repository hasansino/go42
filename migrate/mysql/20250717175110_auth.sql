-- +goose Up

create table if not exists auth_users (
    id bigint unsigned not null auto_increment primary key,
    uuid char(36) not null,
    password varchar(255) null,
    email varchar(255) not null,
    status varchar(255) not null default 'active',
    is_system boolean not null default false,
    metadata json null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null,
    unique key idx_auth_users_uid (uuid),
    unique key idx_auth_users_email (email),
    key idx_auth_users_deleted_at (deleted_at)
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

insert ignore into auth_users (uuid, password, email, is_system) values
('00000000-0000-0000-0000-000000000000', null, 'admin@system.local', true);

create table if not exists auth_roles (
    id bigint unsigned not null auto_increment primary key,
    name varchar(50) not null,
    description varchar(255) not null default '',
    is_system boolean not null default false,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp on update current_timestamp,
    deleted_at timestamp null,
    unique key idx_auth_roles_name (name),
    key idx_auth_roles_deleted_at (deleted_at)
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

insert ignore into auth_roles (name, description, is_system) values
('user', 'standard user role with limited access', 0);

create table if not exists auth_permissions (
    id bigint unsigned not null auto_increment primary key,
    resource varchar(255) not null,
    action varchar(255) not null,
    scope varchar(255) null,
    created_at timestamp not null default current_timestamp,
    unique key idx_auth_permissions_resource_action_scope (resource, action, scope)
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

insert ignore into auth_permissions (resource, action, scope) values
('user', 'read_self', null);

create table if not exists auth_role_permissions (
    role_id bigint unsigned not null,
    permission_id bigint unsigned not null,
    primary key (role_id, permission_id),
    key idx_auth_role_permissions_permission_id (permission_id),
    constraint fk_auth_role_permissions_role_id foreign key (role_id) references auth_roles(id) on delete cascade,
    constraint fk_auth_role_permissions_permission_id foreign key (permission_id) references auth_permissions(id) on delete cascade
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

insert ignore into auth_role_permissions (role_id, permission_id) values
((select id from auth_roles where name = 'user'), (select id from auth_permissions where resource = 'user' and action = 'read_self'));

create table if not exists auth_user_roles (
    user_id bigint unsigned not null,
    role_id bigint unsigned not null,
    granted_at timestamp not null default current_timestamp,
    granted_by bigint unsigned null,
    expires_at timestamp null,
    primary key (user_id, role_id),
    key idx_auth_user_roles_role_id (role_id),
    key idx_auth_user_roles_expires_at (expires_at),
    constraint fk_auth_user_roles_user_id foreign key (user_id) references auth_users(id) on delete cascade,
    constraint fk_auth_user_roles_role_id foreign key (role_id) references auth_roles(id) on delete cascade,
    constraint fk_auth_user_roles_granted_by foreign key (granted_by) references auth_users(id) on delete set null
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS auth_api_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token VARCHAR(255) NOT NULL    scopes JSON DEFAULT NULL,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    UNIQUE KEY idx_token (token),
    CONSTRAINT fk_api_tokens_user FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_token_lookup ON auth_api_tokens(token, deleted_at, expires_at);

-- +goose Down

drop table if exists auth_api_tokens;
drop table if exists auth_user_roles;
drop table if exists auth_role_permissions;
drop table if exists auth_permissions;
drop table if exists auth_roles;
drop table if exists auth_users;
