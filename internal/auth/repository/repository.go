package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/database"
)

type Repository struct {
	*database.BaseRepository
}

func New(baseRepository *database.BaseRepository) *Repository {
	return &Repository{
		BaseRepository: baseRepository,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	err := r.GetTx(ctx).Create(user).Error
	if err != nil {
		if r.IsDuplicateKeyError(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.GetReadDB(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN auth_user_roles ON auth_user_roles.role_id = auth_roles.id").
				Where("auth_user_roles.expires_at IS NULL OR auth_user_roles.expires_at > ?", time.Now())
		}).
		Preload("Roles.Permissions").
		First(&user, "id = ?", id).Error
	if r.IsNotFoundError(err) {
		return nil, domain.ErrEntityNotFound
	}
	return &user, err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.GetReadDB(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN auth_user_roles ON auth_user_roles.role_id = auth_roles.id").
				Where("auth_user_roles.expires_at IS NULL OR auth_user_roles.expires_at > ?", time.Now())
		}).
		Preload("Roles.Permissions").
		First(&user, "email = ?", email).Error
	if r.IsNotFoundError(err) {
		return nil, domain.ErrEntityNotFound
	}
	return &user, err
}
