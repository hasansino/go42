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
	headerAuthorization     = "Authorization"
	headerValueBearerPrefix = "Bearer "
)

type authServiceAccessor interface {
	Logout(ctx context.Context, accessToken, refreshToken string) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	ValidateJWTTokenInternal(ctx context.Context, token string) (*jwt.RegisteredClaims, error)
	InvalidateJWTToken(ctx context.Context, token string, until time.Time) error
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
			case ctx.QueryParam("token") != "":
				// Support JWT token as query parameter for WebSocket connections
				token := ctx.QueryParam("token")
				if err := processUserAuth(ctx, svc, token); err != nil {
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
	claims, err := svc.ValidateJWTTokenInternal(ctx.Request().Context(), token)
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
		UUID: user.UUID.String(),
		Type: domain.AuthenticationTypeCredentials,
	}
	authInfo.SetPermissions(user.PermissionList())

	newCtx := auth.SetAuthToContext(ctx.Request().Context(), authInfo)
	ctx.SetRequest(ctx.Request().WithContext(newCtx))

	return nil
}

// NewJWTOnlyMiddleware creates an auth middleware that only supports JWT tokens (for WebSocket)
func NewJWTOnlyMiddleware(svc authServiceAccessor) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var token string
			var err error

			// Debug logging
			slog.DebugContext(ctx.Request().Context(), "JWT middleware processing request",
				slog.String("path", ctx.Request().URL.Path),
				slog.String("query", ctx.Request().URL.RawQuery),
				slog.String("authorization_header", ctx.Request().Header.Get(headerAuthorization)),
				slog.String("token_query_param", ctx.QueryParam("token")))

			// Try JWT from Authorization header first
			if authHeader := ctx.Request().Header.Get(headerAuthorization); authHeader != "" {
				slog.DebugContext(
					ctx.Request().Context(),
					"Found Authorization header",
					slog.String("header", authHeader),
				)
				token, err = extractBearerToken(authHeader)
				if err != nil {
					slog.ErrorContext(ctx.Request().Context(), "Failed to extract bearer token", slog.Any("error", err))
					return httpAPI.SendJSONError(ctx,
						http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				}
			} else if queryToken := ctx.QueryParam("token"); queryToken != "" {
				// Support JWT token as query parameter for WebSocket connections
				slog.DebugContext(ctx.Request().Context(), "Found token query parameter", slog.String("token_prefix", queryToken[:20]+"..."))
				token = queryToken
			} else {
				// No valid JWT token found
				slog.ErrorContext(ctx.Request().Context(), "No valid JWT token found")
				return httpAPI.SendJSONError(ctx,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			slog.DebugContext(
				ctx.Request().Context(),
				"Processing user auth with token",
				slog.String("token_prefix", token[:20]+"..."),
			)
			if err := processUserAuth(ctx, svc, token); err != nil {
				slog.ErrorContext(ctx.Request().Context(), "User auth failed", slog.Any("error", err))
				return httpAPI.SendJSONError(ctx,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			slog.DebugContext(ctx.Request().Context(), "JWT middleware completed successfully")
			return next(ctx)
		}
	}
}
