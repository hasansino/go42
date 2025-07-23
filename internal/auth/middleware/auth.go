package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/tools"
)

type authServiceAccessor interface {
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	ValidateToken(token string) (*jwt.RegisteredClaims, error)
}

func NewAuthMiddleware(svc authServiceAccessor) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := tools.ExtractBearerTokenFromRequest(c.Request())
			if err != nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			claims, err := svc.ValidateToken(token)
			if err != nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			err = uuid.Validate(claims.Subject)
			if err != nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			user, err := svc.GetUserByUUID(c.Request().Context(), claims.Subject)
			if err != nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			authInfo := &domain.ContextAuthInfo{
				ID:    user.ID,
				UUID:  user.UUID.String(),
				Email: user.Email,
			}
			authInfo.SetRoles(user.RoleList())
			authInfo.SetPermissions(user.PermissionList())

			ctx := auth.SetAuthToContext(c.Request().Context(), authInfo)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
