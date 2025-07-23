-- +goose Up

create table if not exists auth_users (
    id bigserial primary key,
    uuid uuid not null,
    password varchar(255) null,
    email varchar(255) not null,
    status varchar(255) not null default 'active',
    metadata jsonb null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null
);

create unique index if not exists idx_auth_users_uuid on auth_users(uuid);
create unique index if not exists idx_auth_users_email on auth_users(email);
create index if not exists idx_auth_users_deleted_at on auth_users(deleted_at);

create table if not exists auth_roles (
    id bigserial primary key,
    name varchar(50) not null,
    description varchar(255) not null default '',
    is_system boolean not null default false,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp,
    deleted_at timestamp null
);

insert into auth_roles (name, description, is_system) values
('admin', 'administrator role with full access', true),
('user', 'standard user role with limited access', false)
on conflict do nothing;

create unique index if not exists idx_auth_roles_name on auth_roles(name);
create index if not exists idx_auth_roles_deleted_at on auth_roles(deleted_at);

create table if not exists auth_permissions (
    id bigserial primary key,
    resource varchar(255) not null,
    action varchar(255) not null,
    scope varchar(255) null,
    created_at timestamp not null default current_timestamp
);

create unique index if not exists idx_auth_permissions_resource_action_scope on auth_permissions(resource, action, scope);

insert into auth_permissions (resource, action, scope) values
('user', 'read_self', null)
on conflict do nothing;

create table if not exists auth_role_permissions (
    role_id bigint not null,
    permission_id bigint not null,
    primary key (role_id, permission_id),
    constraint fk_auth_role_permissions_role_id foreign key (role_id) references auth_roles(id) on delete cascade,
    constraint fk_auth_role_permissions_permission_id foreign key (permission_id) references auth_permissions(id) on delete cascade
);

insert into auth_role_permissions (role_id, permission_id) values
((select id from auth_roles where name = 'admin'), (select id from auth_permissions)),
((select id from auth_roles where name = 'user'), (select id from auth_permissions where resource = 'user' and action = 'read_self'))
on conflict do nothing;

create table if not exists auth_user_roles (
    user_id bigint not null,
    role_id bigint not null,
    granted_at timestamp not null default current_timestamp,
    granted_by bigint null,
    expires_at timestamp null,
    primary key (user_id, role_id),
    constraint fk_auth_user_roles_user_id foreign key (user_id) references auth_users(id) on delete cascade,
    constraint fk_auth_user_roles_role_id foreign key (role_id) references auth_roles(id) on delete cascade,
    constraint fk_auth_user_roles_granted_by foreign key (granted_by) references auth_users(id) on delete set null
);

create index if not exists idx_auth_user_roles_role_id on auth_user_roles(role_id);
create index if not exists idx_auth_user_roles_expires_at on auth_user_roles(expires_at);
create index if not exists idx_auth_role_permissions_permission_id on auth_role_permissions(permission_id);

-- +goose Down

drop table if exists auth_user_roles;
drop table if exists auth_role_permissions;
drop table if exists auth_permissions;
drop table if exists auth_roles;
drop table if exists auth_users;
