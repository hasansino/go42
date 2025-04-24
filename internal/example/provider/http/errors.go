package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/hasansino/goapp/internal/api"
	"github.com/hasansino/goapp/internal/example"
)

func (h *Handler) processError(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, example.ErrExample):
		return api.SendJSONError(ctx, http.StatusTeapot, err.Error())
	default:
		return api.SendJSONError(ctx, http.StatusInternalServerError, err.Error())
	}
}
