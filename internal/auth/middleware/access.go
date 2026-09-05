package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"

	httpAPI "github.com/go42-dev/go42/internal/api/http"
	"github.com/go42-dev/go42/internal/auth"
)

func NewAccessMiddleware(permissions ...string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authInfo := auth.RetrieveAuthFromContext(c.Request().Context())
			if authInfo == nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}
			for _, permission := range permissions {
				if !authInfo.HasPermission(permission) {
					return httpAPI.SendJSONError(c,
						http.StatusForbidden, http.StatusText(http.StatusForbidden))
				}
			}
			return next(c)
		}
	}
}
