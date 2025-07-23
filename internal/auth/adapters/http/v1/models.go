package adapter

import "github.com/hasansino/go42/internal/auth/models"

type UserResponse struct {
	UUID        string   `json:"uuid"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func UserResponseFromModel(user *models.User) UserResponse {
	return UserResponse{
		UUID:        user.UUID.String(),
		Roles:       user.RoleList(),
		Permissions: user.PermissionList(),
	}
}
