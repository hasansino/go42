-- +goose Up

insert or ignore into auth_users (uuid, password, email, is_system) values
('00000000-0000-0000-0000-000000000000', null, 'admin@system.local', 1);

insert or ignore into auth_roles (name, description, is_system) values
('admin', 'full access to entire system', 1),
('user', 'standard user role with limited access', 0);

insert or ignore into auth_permissions (resource, action) values
('users', 'read_self'),
('users', 'update_self'),
('users', 'list'),
('users', 'read_others'),
('users', 'create'),
('users', 'update'),
('users', 'update');

-- admins have all permissions
insert or ignore into auth_role_permissions (role_id, permission_id)
select
    (
        select auth_roles.id
        from auth_roles
        where auth_roles.name = 'admin'
    ) as role_id,
    auth_permissions.id as permission_id
from
    auth_permissions;

-- users can read & update themselves
insert or ignore into auth_role_permissions (role_id, permission_id) values
(
    (
        select auth_roles.id
        from auth_roles
        where auth_roles.name = 'user'
    ),
    (
        select auth_permissions.id
        from
            auth_permissions
        where
            auth_permissions.resource = 'users'
            and auth_permissions.action = 'read_self'
    )
),
(
    (
        select auth_roles.id
        from auth_roles
        where auth_roles.name = 'user'
    ),
    (
        select auth_permissions.id
        from auth_permissions
        where
            auth_permissions.resource = 'users'
            and auth_permissions.action = 'update_self'
    )
);

-- +goose Down

delete from auth_role_permissions;
delete from auth_permissions;
delete from auth_roles;
delete from auth_users;
