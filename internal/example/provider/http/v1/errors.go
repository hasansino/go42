package provider

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/example/domain"
)

func (p *Provider) processError(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return httpAPI.SendJSONError(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return httpAPI.SendJSONError(ctx, http.StatusConflict, err.Error())
	default:
		return httpAPI.SendJSONError(ctx, http.StatusInternalServerError, err.Error())
	}
}
