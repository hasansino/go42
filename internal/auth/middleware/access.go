package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/auth"
)

//go:generate mockgen -destination=./mocks/mocks.go -package=mocks .

type accessServiceAccessor interface {
}

func NewAccessMiddleware(_ accessServiceAccessor) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authInfo := auth.RetrieveAuthFromContext(c.Request().Context())
			if authInfo == nil {
				return httpAPI.SendJSONError(c,
					http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}
			return next(c)
		}
	}
}
