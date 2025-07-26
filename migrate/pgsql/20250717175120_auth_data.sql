-- +goose Up

insert into auth_users (uuid, password, email, is_system) values
('00000000-0000-0000-0000-000000000000', null, 'admin@system.local', true)
on conflict do nothing;

insert into auth_roles (name, description, is_system) values
('user', 'standard user role with limited access', false)
on conflict do nothing;

insert into auth_permissions (resource, action, scope) values
('user', 'read_self', null)
on conflict do nothing;

insert into auth_role_permissions (role_id, permission_id) values
((select id from auth_roles where name = 'user'), (select id from auth_permissions where resource = 'user' and action = 'read_self'))
on conflict do nothing;

-- +goose Down

truncate table auth_role_permissions;
truncate table auth_permissions;
truncate table auth_roles;
truncate table auth_users;
