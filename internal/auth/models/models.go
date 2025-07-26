package models

import (
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hasansino/go42/internal/auth/domain"
)

type User struct {
	ID        int
	UUID      uuid.UUID
	Email     string
	Password  sql.Null[string]
	Status    string
	IsSystem  bool
	Metadata  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Roles []Role `gorm:"many2many:auth_user_roles;association_autoupdate:false"`
}

func (*User) TableName() string { return "auth_users" }

func (u *User) IsActive() bool {
	return u.Status == domain.UserStatusActive
}

func (u *User) RoleList() []string {
	roles := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roles[i] = role.Name
	}
	return roles
}

func (u *User) PermissionList() []string {
	permissions := make([]string, 0)
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if !slices.Contains(permissions, permission.Resource+":"+permission.Action) {
				permissions = append(permissions, permission.Resource+":"+permission.Action)
			}
		}
	}
	return permissions
}

type UserHistoryRecord struct {
	ID         uuid.UUID
	OccurredAt time.Time
	CreatedAt  time.Time
	UserID     int
	EventType  string
	Data       []byte
	Metadata   string
}

func (*UserHistoryRecord) TableName() string { return "auth_users_history" }

type Role struct {
	ID          int
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt

	Permissions []Permission `gorm:"many2many:auth_role_permissions;"`
}

func (Role) TableName() string { return "auth_roles" }

type Permission struct {
	ID        int
	Resource  string
	Action    string
	Scope     sql.Null[string]
	CreatedAt time.Time
}

func (Permission) TableName() string { return "auth_permissions" }

type UserRole struct {
	UserID    int
	RoleID    int
	GrantedAt time.Time
	GrantedBy sql.Null[int]
	ExpiresAt sql.Null[time.Time]
}

func (UserRole) TableName() string { return "auth_user_roles" }

type Token struct {
	ID         int
	UserID     int
	Token      string
	Name       string
	LastUsedAt sql.Null[time.Time]
	ExpiresAt  sql.Null[time.Time]
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt

	Permissions []Permission `gorm:"many2many:auth_api_tokens_permissions;"`
}

func (*Token) TableName() string { return "auth_api_tokens" }

func (a *Token) PermissionList() []string {
	permissions := make([]string, 0)
	for _, permission := range a.Permissions {
		if !slices.Contains(permissions, permission.Resource+":"+permission.Action) {
			permissions = append(permissions, permission.Resource+":"+permission.Action)
		}
	}
	return permissions
}
