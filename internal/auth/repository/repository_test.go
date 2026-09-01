package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	authRepository "github.com/hasansino/go42/internal/auth/repository"
	"github.com/hasansino/go42/internal/cache"
	cacheLocal "github.com/hasansino/go42/internal/cache/local"
	"github.com/hasansino/go42/internal/database"
	databaseSqlite "github.com/hasansino/go42/internal/database/sqlite"
)

func TestRepository_UserCacheBypassesTransactionsAndInvalidatesAliases(t *testing.T) {
	dbEngine, err := databaseSqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	sqlDB, err := dbEngine.Master().DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := dbEngine.Master()
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}

	user := &models.User{
		UUID:   uuid.New(),
		Email:  "old@example.com",
		Status: domain.UserStatusActive,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	cacheEngine := cacheLocal.New()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = cacheEngine.Shutdown(shutdownCtx)
	})
	repository := authRepository.New(
		database.NewBaseRepository(dbEngine),
		cacheEngine,
		time.Minute,
		time.Minute,
	)
	ctx := context.Background()

	// Prime all three independent lookup aliases with the old record.
	if _, err := repository.GetUserByID(ctx, user.ID); err != nil {
		t.Fatalf("prime ID cache: %v", err)
	}
	if _, err := repository.GetUserByUUID(ctx, user.UUID.String()); err != nil {
		t.Fatalf("prime UUID cache: %v", err)
	}
	if _, err := repository.GetUserByEmail(ctx, user.Email); err != nil {
		t.Fatalf("prime email cache: %v", err)
	}

	const newEmail = "new@example.com"
	if err := db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("email", newEmail).Error; err != nil {
		t.Fatalf("update database directly: %v", err)
	}

	// A transaction must see database state, not the previously cached record.
	var transactionalUser *models.User
	if err := repository.WithTransaction(ctx, func(txCtx context.Context) error {
		var lookupErr error
		transactionalUser, lookupErr = repository.GetUserByUUID(txCtx, user.UUID.String())
		return lookupErr
	}); err != nil {
		t.Fatalf("transactional lookup: %v", err)
	}
	if transactionalUser.Email != newEmail {
		t.Errorf("transactional email = %q, want %q", transactionalUser.Email, newEmail)
	}

	if err := repository.InvalidateUserCache(
		ctx, user.ID, user.UUID.String(), user.Email, newEmail,
	); err != nil {
		t.Fatalf("invalidate user cache: %v", err)
	}

	byID, err := repository.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("lookup by ID after invalidation: %v", err)
	}
	if byID.Email != newEmail {
		t.Errorf("email by ID = %q, want %q", byID.Email, newEmail)
	}
	byUUID, err := repository.GetUserByUUID(ctx, user.UUID.String())
	if err != nil {
		t.Fatalf("lookup by UUID after invalidation: %v", err)
	}
	if byUUID.Email != newEmail {
		t.Errorf("email by UUID = %q, want %q", byUUID.Email, newEmail)
	}
	if _, err := repository.GetUserByEmail(ctx, user.Email); !errors.Is(err, domain.ErrEntityNotFound) {
		t.Errorf("old email lookup error = %v, want %v", err, domain.ErrEntityNotFound)
	}
	if _, err := repository.GetUserByEmail(ctx, newEmail); err != nil {
		t.Fatalf("lookup by new email after invalidation: %v", err)
	}
}

func TestRepository_GetTokenLoadsPermissions(t *testing.T) {
	dbEngine, err := databaseSqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	sqlDB, err := dbEngine.Master().DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := dbEngine.Master()
	if err := db.AutoMigrate(&models.Token{}, &models.Permission{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE auth_api_tokens_permissions (
			token_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			PRIMARY KEY (token_id, permission_id)
		)
	`).Error; err != nil {
		t.Fatalf("create token permissions table: %v", err)
	}

	permission := &models.Permission{Resource: "users", Action: "read_self"}
	if err := db.Create(permission).Error; err != nil {
		t.Fatalf("create permission: %v", err)
	}
	apiToken := &models.Token{
		UUID:   uuid.New(),
		UserID: 42,
		Token:  "hashed-api-token",
		Name:   "test token",
	}
	if err := db.Create(apiToken).Error; err != nil {
		t.Fatalf("create API token: %v", err)
	}
	if err := db.Exec(
		"INSERT INTO auth_api_tokens_permissions (token_id, permission_id) VALUES (?, ?)",
		apiToken.ID,
		permission.ID,
	).Error; err != nil {
		t.Fatalf("assign permission: %v", err)
	}

	repository := authRepository.New(
		database.NewBaseRepository(dbEngine),
		cache.NewNoop(),
		time.Minute,
		time.Minute,
	)
	got, err := repository.GetToken(context.Background(), apiToken.Token)
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}
	if got.ID != apiToken.ID {
		t.Errorf("GetToken() ID = %d, want %d", got.ID, apiToken.ID)
	}
	permissions := got.PermissionList()
	if len(permissions) != 1 || permissions[0] != "users:read_self" {
		t.Errorf("GetToken() permissions = %v, want [users:read_self]", permissions)
	}
}
