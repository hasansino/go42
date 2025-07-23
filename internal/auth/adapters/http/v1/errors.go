package adapter

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/auth/domain"
)

func (a *Adapter) processError(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrEntityNotFound):
		return httpAPI.SendJSONError(ctx,
			http.StatusNotFound, http.StatusText(http.StatusNotFound))
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return httpAPI.SendJSONError(ctx,
			http.StatusConflict, http.StatusText(http.StatusConflict))
	case errors.Is(err, domain.ErrInvalidCredentials):
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	case errors.Is(err, domain.ErrUserInactive):
		return httpAPI.SendJSONError(ctx,
			http.StatusForbidden, http.StatusText(http.StatusForbidden))
	case errors.Is(err, domain.ErrInvalidToken):
		return httpAPI.SendJSONError(ctx,
			http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	case errors.Is(err, domain.ErrTokenExpired):
		return httpAPI.SendJSONError(ctx,
			http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	default:
		return httpAPI.SendJSONError(ctx,
			http.StatusInternalServerError, err.Error())
	}
}
