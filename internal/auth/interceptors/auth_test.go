package interceptors_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	authInterceptors "github.com/hasansino/go42/internal/auth/interceptors"
	authMocks "github.com/hasansino/go42/internal/auth/mocks"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/cache"
)

func TestUnaryAuthInterceptor_APITokenUsesOwnerIdentityAndTokenPermissions(t *testing.T) {
	const rawToken = "api_kXqdf2uQ7hmOARp-pZrhA6_IsZSeKCmSEM4YFKBGIzA"
	user := &models.User{
		ID:     42,
		UUID:   uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Status: domain.UserStatusActive,
		Roles: []models.Role{{
			Permissions: []models.Permission{{Resource: "users", Action: "delete"}},
		}},
	}
	apiToken := &models.Token{
		ID:     7,
		UUID:   uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
		UserID: user.ID,
		Permissions: []models.Permission{{
			Resource: "users",
			Action:   "list",
		}},
	}

	ctrl := gomock.NewController(t)
	repository := authMocks.NewMockrepository(ctrl)
	repository.EXPECT().GetToken(gomock.Any(), sha256String(rawToken)).Return(apiToken, nil)
	repository.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)
	service := auth.NewService(
		repository,
		authMocks.NewMockoutboxService(ctrl),
		cache.NewNoop(),
	)

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-api-key", rawToken),
	)
	interceptor := authInterceptors.NewUnaryAuthInterceptor(service)
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/ListUsers"},
		func(ctx context.Context, _ interface{}) (interface{}, error) {
			authInfo := auth.RetrieveAuthFromContext(ctx)
			if authInfo == nil {
				t.Fatal("authentication context is nil")
			}
			if authInfo.ID != user.ID || authInfo.UUID != user.UUID.String() {
				t.Errorf(
					"authentication identity = (%d, %q), want (%d, %q)",
					authInfo.ID, authInfo.UUID, user.ID, user.UUID.String(),
				)
			}
			if authInfo.Type != domain.AuthenticationTypeApiToken {
				t.Errorf("authentication type = %q, want %q", authInfo.Type, domain.AuthenticationTypeApiToken)
			}
			if !authInfo.HasPermission(domain.RBACPermissionUsersList) {
				t.Error("API-token permission is missing")
			}
			if authInfo.HasPermission(domain.RBACPermissionUsersDelete) {
				t.Error("owner role permission leaked into API-token permissions")
			}
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor error = %v", err)
	}
}

func sha256String(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}
