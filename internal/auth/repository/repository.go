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
	return r.getUser(ctx, map[string]any{"id": id})
}

func (r *Repository) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return r.getUser(ctx, map[string]any{"uuid": uuid})
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.getUser(ctx, map[string]any{"email": email})
}

func (r *Repository) getUser(ctx context.Context, filter map[string]any) (*models.User, error) {
	var user models.User
	tx := r.GetReadDB(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN auth_user_roles ON auth_user_roles.user_id = id").
				Where("auth_user_roles.expires_at IS NULL OR auth_user_roles.expires_at > ?", time.Now())
		}).
		Preload("Roles.Permissions")

	for key, value := range filter {
		tx = tx.Where(fmt.Sprintf("%s = ?", key), value)
	}

	err := tx.First(&user).Error

	if r.IsNotFoundError(err) {
		return nil, domain.ErrEntityNotFound
	}

	return &user, err
}

func (r *Repository) AssignRoleToUser(ctx context.Context, userID int, role string) error {
	var roleID int
	err := r.GetTx(ctx).Raw("SELECT id FROM auth_roles WHERE name = ?", role).Scan(&roleID).Error
	if err != nil {
		return fmt.Errorf("error finding role: %w", err)
	}
	err = r.GetTx(ctx).Exec(
		"INSERT INTO auth_user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		userID, roleID,
	).Error
	if err != nil {
		return fmt.Errorf("error assigning role to user: %w", err)
	}
	return nil
}
