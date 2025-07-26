package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *Repository) AssignRoleToUser(ctx context.Context, userID int, roleName string) error {
	var role models.Role
	err := r.GetTx(ctx).
		Where("name = ?", roleName).
		First(&role).Error
	if err != nil {
		return fmt.Errorf("error retrieving role: %w", err)
	}

	userRole := models.UserRole{
		UserID: userID,
		RoleID: role.ID,
	}

	err = r.GetTx(ctx).
		Clauses(clause.Insert{Modifier: "IGNORE"}).
		Create(&userRole).Error
	if err != nil {
		return fmt.Errorf("error assigning role to user: %w", err)
	}

	return nil
}

func (r *Repository) GetToken(ctx context.Context, token string) (*models.Token, error) {
	var apiToken models.Token
	err := r.GetReadDB(ctx).
		Preload("Permissions").
		Where("token = ?", token).
		First(&apiToken).Error

	if r.IsNotFoundError(err) {
		return nil, domain.ErrEntityNotFound
	}

	return &apiToken, err
}

func (r *Repository) UpdateTokenLastUsed(ctx context.Context, tokenID int, when time.Time) error {
	result := r.GetTx(ctx).
		Model(&models.Token{}).
		Where("id = ?", tokenID).
		Update("last_used_at", when)

	if r.IsNotFoundError(result.Error) || result.RowsAffected == 0 {
		return domain.ErrEntityNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error updating api token: %w", result.Error)
	}
	return nil
}

func (r *Repository) SaveUserHistoryRecord(ctx context.Context, record *models.UserHistoryRecord) error {
	return r.GetTx(ctx).Create(record).Error
}
