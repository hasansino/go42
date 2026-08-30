package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"

	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/middleware"
	authMocks "github.com/hasansino/go42/internal/auth/mocks"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/cache"
)

func TestAuthMiddleware_RequiresAccessToken(t *testing.T) {
	const password = "correct horse battery staple"
	user := &models.User{
		ID:     42,
		UUID:   uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Email:  "alice@example.com",
		Status: domain.UserStatusActive,
	}
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("set test password: %v", err)
	}

	ctrl := gomock.NewController(t)
	repository := authMocks.NewMockrepository(ctrl)
	repository.EXPECT().
		GetUserByEmail(gomock.Any(), user.Email).
		Return(user, nil)
	repository.EXPECT().
		GetUserByUUID(gomock.Any(), user.UUID.String()).
		Return(user, nil).
		AnyTimes()

	outbox := authMocks.NewMockoutboxService(ctrl)
	outbox.EXPECT().
		NewOutboxMessage(gomock.Any(), domain.TopicNameAuthEvents, gomock.Any()).
		Return(nil)

	service := auth.NewService(
		repository,
		outbox,
		cache.NewNoop(),
		auth.WithJWTSecrets([]string{"token-purpose-test-secret"}),
		auth.WithJWTAccessTokenTTL(15*time.Minute),
		auth.WithJWTRefreshTokenTTL(24*time.Hour),
		auth.WithJWTIssuer("go42-test"),
		auth.WithJWTAudience([]string{"go42-test"}),
	)
	tokens, err := service.Login(context.Background(), user.Email, password)
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	var handlerCalls atomic.Int32
	e := echo.New()
	e.GET("/protected", func(c *echo.Context) error {
		handlerCalls.Add(1)
		return c.NoContent(http.StatusNoContent)
	}, middleware.NewAuthMiddleware(service))

	accessRecorder := performAuthenticatedRequest(e, tokens.AccessToken)
	if accessRecorder.Code != http.StatusNoContent {
		t.Fatalf("access token status = %d, want %d", accessRecorder.Code, http.StatusNoContent)
	}

	refreshRecorder := performAuthenticatedRequest(e, tokens.RefreshToken)
	if refreshRecorder.Code != http.StatusUnauthorized {
		t.Errorf("refresh token status = %d, want %d", refreshRecorder.Code, http.StatusUnauthorized)
	}
	if got := handlerCalls.Load(); got != 1 {
		t.Errorf("protected handler calls = %d, want 1", got)
	}
}

func performAuthenticatedRequest(e *echo.Echo, token string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	e.ServeHTTP(recorder, request)
	return recorder
}
