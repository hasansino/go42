package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hasansino/go42/internal/auth/models"
	authRepository "github.com/hasansino/go42/internal/auth/repository"
	"github.com/hasansino/go42/internal/cache"
	"github.com/hasansino/go42/internal/database"
	databaseSqlite "github.com/hasansino/go42/internal/database/sqlite"
)

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
