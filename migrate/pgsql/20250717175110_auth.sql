-- +goose Up

CREATE TABLE IF NOT EXISTS auth_users (
    id BIGSERIAL PRIMARY KEY,
    uuid uuid NOT NULL,
    password VARCHAR(255) NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'active',
    metadata JSONB NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_uuid ON auth_users(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_email ON auth_users(email);
CREATE INDEX IF NOT EXISTS idx_auth_users_deleted_at ON auth_users(deleted_at);

--

CREATE TABLE IF NOT EXISTS auth_roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_roles_name ON auth_roles(name);
CREATE INDEX IF NOT EXISTS idx_auth_roles_deleted_at ON auth_roles(deleted_at);

--

CREATE TABLE IF NOT EXISTS auth_permissions (
    id BIGSERIAL PRIMARY KEY,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    scope VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_permissions_resource_action_scope ON auth_permissions(resource, action, scope);

--

CREATE TABLE IF NOT EXISTS auth_role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_auth_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES auth_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES auth_permissions(id) ON DELETE CASCADE
);

--

CREATE TABLE IF NOT EXISTS auth_user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by BIGINT NULL,
    expires_at TIMESTAMP NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_auth_user_roles_user_id FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_roles_role_id FOREIGN KEY (role_id) REFERENCES auth_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_roles_granted_by FOREIGN KEY (granted_by) REFERENCES auth_users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_user_roles_role_id ON auth_user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_auth_user_roles_expires_at ON auth_user_roles(expires_at);

CREATE INDEX IF NOT EXISTS idx_auth_role_permissions_permission_id ON auth_role_permissions(permission_id);

-- +goose Down

DROP TABLE IF EXISTS auth_user_roles;
DROP TABLE IF EXISTS auth_role_permissions;
DROP TABLE IF EXISTS auth_permissions;
DROP TABLE IF EXISTS auth_roles;
DROP TABLE IF EXISTS auth_users;
