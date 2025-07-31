package domain

import (
	"errors"
	"time"
)

// RBAC roles.
// MUST reflect state of table `auth_roles`.

const (
	RBACRoleUser = "user"
)

var RBACAllRoles = []string{
	RBACRoleUser,
}

// RBAC permissions.
// MUST reflect state of table `auth_permissions`.

const (
	RBACPermissionUsersReadSelf   = "users:read_self"
	RBACPermissionUsersUpdateSelf = "users:update_self"
	RBACPermissionUsersList       = "users:list"
	RBACPermissionUsersReadOthers = "users:read_others"
	RBACPermissionUsersCreate     = "users:create"
	RBACPermissionUsersUpdate     = "users:update"
	RBACPermissionUsersDelete     = "users:delete"
)

var RBACAllPermissions = []string{
	RBACPermissionUsersReadSelf,
	RBACPermissionUsersUpdateSelf,
	RBACPermissionUsersList,
	RBACPermissionUsersReadOthers,
	RBACPermissionUsersCreate,
	RBACPermissionUsersUpdate,
	RBACPermissionUsersDelete,
}

// ---- RBAC END

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

var UserStatuses = []string{
	UserStatusActive,
	UserStatusInactive,
}

const (
	TopicNameAuthEvents = "auth_events_topic"
	EventTypeAuthSignUp = "auth.signup"
	EventTypeAuthLogin  = "auth.login"
	EventTypeAuthLogout = "auth.logout"
	EventTypeUserCreate = "user.create"
	EventTypeUserUpdate = "user.update"
	EventTypeUserDelete = "user.delete"
)

var (
	ErrEntityNotFound     = errors.New("entity not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrPasswordWeak       = errors.New("password is too weak")
)

// ----

// Tokens represents the structure of JWT authentication tokens.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// TokenWasUsed is a message passed to channel for outside consumers.
type TokenWasUsed struct {
	ID   int
	When time.Time
}

// ----

type AuthenticationType string

const (
	AuthenticationTypeCredentials = "credentials"
	AuthenticationTypeApiToken    = "api_token"
)

// ContextAuthInfo holds authentication information in the request context.
// ID field stores authenticated subject id which is described by Type field.
type ContextAuthInfo struct {
	ID            int
	UUID          string
	Type          AuthenticationType
	permissions   []string
	permissionMap map[string]struct{}
}

func (ctx *ContextAuthInfo) SetPermissions(permissions []string) {
	ctx.permissions = permissions
	ctx.permissionMap = make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		ctx.permissionMap[permission] = struct{}{}
	}
}

func (ctx *ContextAuthInfo) HasPermission(permission string) bool {
	if ctx.permissionMap == nil {
		return false
	}
	_, exists := ctx.permissionMap[permission]
	return exists
}

// ----

type CreateUserData struct {
	Email    string
	Password string
}

type UpdateUserData struct {
	Email    *string
	Password *string
}
