package domain

import (
	"errors"
)

// RBAC roles.
// MUST reflect state of table `auth_roles`.

const (
	RBACRoleUser = "user"
)

// RBAC permissions.
// MUST reflect state of table `auth_permissions`.

const (
	RBACPermissionUserReadSelf = "user:read_self"
)

const (
	UserStatusActive = "active"
)

const (
	TopicNameAuthEvents = "auth_events_topic"
	EventTypeSignUp     = "auth.signup"
	EventTypeLogin      = "auth.login"
)

var (
	ErrEntityNotFound     = errors.New("entity not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// Tokens represents the structure of JWT authentication tokens.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthenticationType string

const (
	AuthenticationTypeCredentials = "user"
	AuthenticationTypeApiToken    = "api_token"
)

// ContextAuthInfo holds authentication information in the request context.
// ID field stores authenticated subject id which is described by Type field.
type ContextAuthInfo struct {
	ID            int
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
