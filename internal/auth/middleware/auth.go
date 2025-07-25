package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

const (
	headerXApiToken         = "x-api-key"
	headerAuthorization     = "Authorization"
	headerValueBearerPrefix = "Bearer "
)

type authServiceAccessor interface {
	Logout(ctx context.Context, accessToken, refreshToken string) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	ValidateJWTToken(ctx context.Context, token string) (*jwt.RegisteredClaims, error)
	InvalidateJWTToken(ctx context.Context, token string, until time.Time) error
	ValidateAPIToken(ctx context.Context, token string) (*models.Token, error)
}

func NewAuthMiddleware(svc authServiceAccessor) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			switch {
			case ctx.Request().Header.Get(headerAuthorization) != "":
				token, err := extractBearerToken(
					ctx.Request().Header.Get(headerAuthorization),
				)
				if err != nil {
					return httpAPI.SendJSONError(ctx,
						http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
				if err := processUserAuth(ctx, svc, token); err != nil {
					return httpAPI.SendJSONError(ctx,
						http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
			case ctx.Request().Header.Get(headerXApiToken) != "":
				if err := processTokenAuth(
					ctx, svc, ctx.Request().Header.Get(headerXApiToken),
				); err != nil {
					return httpAPI.SendJSONError(ctx,
						http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
			default:
				return httpAPI.SendJSONError(ctx,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			return next(ctx)
		}
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, headerValueBearerPrefix) {
		return "", fmt.Errorf("invalid Authorization header format, expected Bearer token")
	}
	token := authHeader[len(headerValueBearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("empty token in Authorization header")
	}
	return token, nil
}

func processUserAuth(ctx echo.Context, svc authServiceAccessor, token string) error {
	claims, err := svc.ValidateJWTToken(ctx.Request().Context(), token)
	if err != nil {
		return fmt.Errorf("invalid access token: %w", err)
	}

	err = uuid.Validate(claims.Subject)
	if err != nil {
		return fmt.Errorf("access token is not valid uuid: %w", err)
	}

	user, err := svc.GetUserByUUID(ctx.Request().Context(), claims.Subject)
	if err != nil {
		return fmt.Errorf("error retrieveing user: %w", err)
	}

	if !user.IsActive() {
		if err := svc.InvalidateJWTToken(ctx.Request().Context(), token, claims.ExpiresAt.Time); err != nil {
			slog.ErrorContext(
				ctx.Request().Context(), "failed to invalidate access token",
				slog.String("component", "auth-access-middleware"),
				slog.Any("error", err),
			)
		}
		return errors.New("user is not allowed to authenticate")
	}

	authInfo := domain.ContextAuthInfo{
		ID:   user.ID,
		Type: domain.AuthenticationTypeCredentials,
	}
	authInfo.SetPermissions(user.PermissionList())

	newCtx := auth.SetAuthToContext(ctx.Request().Context(), authInfo)
	ctx.SetRequest(ctx.Request().WithContext(newCtx))

	return nil
}

func processTokenAuth(ctx echo.Context, svc authServiceAccessor, token string) error {
	apiToken, err := svc.ValidateAPIToken(ctx.Request().Context(), token)
	if err != nil {
		return fmt.Errorf("invalid access token: %w", err)
	}

	user, err := svc.GetUserByID(ctx.Request().Context(), apiToken.UserID)
	if err != nil {
		return fmt.Errorf("error retrieveing user: %w", err)
	}

	if !user.IsActive() {
		return errors.New("user is not allowed to authenticate")
	}

	authInfo := domain.ContextAuthInfo{
		ID:   apiToken.ID,
		Type: domain.AuthenticationTypeApiToken,
	}

	// for api token auth we are using only token permissions
	authInfo.SetPermissions(apiToken.PermissionList())

	newCtx := auth.SetAuthToContext(ctx.Request().Context(), authInfo)
	ctx.SetRequest(ctx.Request().WithContext(newCtx))

	return nil
}
