package domain

import (
	"errors"
)

const (
	UserStatusActive = "active"
)

var (
	ErrEntityNotFound     = errors.New("entity not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is not active")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

// Tokens represents the structure of JWT authentication tokens.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// ContextAuthInfo holds authentication information in the request context.
type ContextAuthInfo struct {
	ID            int
	UUID          string
	Email         string
	roles         []string
	roleMap       map[string]struct{}
	permissions   []string
	permissionMap map[string]struct{}
}

func (ctx *ContextAuthInfo) SetRoles(roles []string) {
	ctx.roles = roles
	ctx.roleMap = make(map[string]struct{}, len(roles))
	for _, role := range roles {
		ctx.roleMap[role] = struct{}{}
	}
}

func (ctx *ContextAuthInfo) SetPermissions(permissions []string) {
	ctx.permissions = permissions
	ctx.permissionMap = make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		ctx.permissionMap[permission] = struct{}{}
	}
}

func (ctx *ContextAuthInfo) HasRole(role string) bool {
	if ctx.roleMap == nil {
		return false
	}
	_, exists := ctx.roleMap[role]
	return exists
}

func (ctx *ContextAuthInfo) HasPermission(permission string) bool {
	if ctx.permissionMap == nil {
		return false
	}
	_, exists := ctx.permissionMap[permission]
	return exists
}
