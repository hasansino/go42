package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int              `json:"-"`
	UID       string           `json:"uid"`
	Email     string           `json:"email"`
	Password  sql.Null[string] `json:"-"`
	Status    string           `json:"status"`
	Metadata  json.RawMessage  `json:"-"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
	DeletedAt gorm.DeletedAt   `json:"-"`

	Roles []Role `gorm:"many2many:auth_user_roles" json:"roles,omitempty"`
}

func (User) TableName() string { return "auth_users" }

type Role struct {
	ID          int            `json:"-"`
	Name        string         `json:"name"`
	Description string         `json:"-"`
	IsSystem    bool           `json:"-"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `json:"-"`

	Permissions []Permission `gorm:"many2many:auth_role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "auth_roles" }

type Permission struct {
	ID        int              `json:"-"`
	Resource  string           `json:"resource"`
	Action    string           `json:"action"`
	Scope     sql.Null[string] `json:"scope"`
	CreatedAt time.Time        `json:"-"`
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
