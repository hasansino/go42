package adapter

import (
	"time"

	"github.com/hasansino/go42/internal/auth/models"
)

type UserResponse struct {
	UUID        string   `json:"uuid"`
	Email       string   `json:"email"`
	CreatedAt   string   `json:"created_at"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func UserResponseFromModel(user *models.User) UserResponse {
	return UserResponse{
		UUID:        user.UUID.String(),
		Email:       user.Email,
		CreatedAt:   user.CreatedAt.Format(time.DateTime),
		Roles:       user.RoleList(),
		Permissions: user.PermissionList(),
	}
}
