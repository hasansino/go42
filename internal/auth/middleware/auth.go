package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/tools"
)

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks .

type authServiceAccessor interface {
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	ValidateToken(token string) (*domain.JWTClaims, error)
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

			user, err := svc.GetUserByID(c.Request().Context(), claims.UserID)
			if err != nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}

			roles := make([]string, len(user.Roles))
			for i, role := range user.Roles {
				roles[i] = role.Name
			}

			authInfo := &domain.ContextAuthInfo{
				UserID: user.ID,
				Email:  user.Email,
				Roles:  roles,
			}

			ctx := auth.SetAuthToContext(c.Request().Context(), authInfo)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
