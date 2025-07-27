-- +goose Up

insert ignore into auth_users (uuid, password, email, is_system) values
('00000000-0000-0000-0000-000000000000', null, 'admin@system.local', true);

insert ignore into auth_roles (name, description, is_system) values
('admin', 'full access to entire system', true),
('user', 'standard user role with limited access', false);

insert ignore into auth_permissions (resource, action) values
('users', 'read_self'),
('users', 'update_self'),
('users', 'list'),
('users', 'read_others'),
('users', 'create'),
('users', 'update'),
('users', 'update');

-- admins have all permissions
insert ignore into auth_role_permissions (role_id, permission_id)
select
    (select id from auth_roles where name = 'admin'), ap.id
from auth_permissions ap;

-- users can read & update themselves
insert ignore into auth_role_permissions (role_id, permission_id) values
(
    (select id from auth_roles where name = 'user'),
    (select id from auth_permissions where resource = 'users' and action = 'read_self')
),
(
    (select id from auth_roles where name = 'user'),
    (select id from auth_permissions where resource = 'users' and action = 'update_self')
);

-- +goose Down

truncate table auth_role_permissions;
truncate table auth_permissions;
truncate table auth_roles;
truncate table auth_users;
