-- +goose Up

CREATE TABLE IF NOT EXISTS auth_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    uid char(36) NOT NULL,
    password VARCHAR(255) NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'active',
    metadata JSON NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_auth_users_uid (uid),
    UNIQUE KEY idx_auth_users_email (email),
    KEY idx_auth_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--

CREATE TABLE IF NOT EXISTS auth_roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY idx_auth_roles_name (name),
    KEY idx_auth_roles_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--

CREATE TABLE IF NOT EXISTS auth_permissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    scope VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY idx_auth_permissions_resource_action_scope (resource, action, scope)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--

CREATE TABLE IF NOT EXISTS auth_role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    KEY idx_auth_role_permissions_permission_id (permission_id),
    CONSTRAINT fk_auth_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES auth_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES auth_permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--

CREATE TABLE IF NOT EXISTS auth_user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by BIGINT UNSIGNED NULL,
    expires_at TIMESTAMP NULL,
    PRIMARY KEY (user_id, role_id),
    KEY idx_auth_user_roles_role_id (role_id),
    KEY idx_auth_user_roles_expires_at (expires_at),
    CONSTRAINT fk_auth_user_roles_user_id FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_roles_role_id FOREIGN KEY (role_id) REFERENCES auth_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_roles_granted_by FOREIGN KEY (granted_by) REFERENCES auth_users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down

DROP TABLE IF EXISTS auth_user_roles;
DROP TABLE IF EXISTS auth_role_permissions;
DROP TABLE IF EXISTS auth_permissions;
DROP TABLE IF EXISTS auth_roles;
DROP TABLE IF EXISTS auth_users;