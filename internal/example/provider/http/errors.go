package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/hasansino/goapp/internal/api"
	"github.com/hasansino/goapp/internal/example/domain"
)

func (h *Handler) processError(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return api.SendJSONError(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return api.SendJSONError(ctx, http.StatusConflict, err.Error())
	default:
		return api.SendJSONError(ctx, http.StatusInternalServerError, err.Error())
	}
}
