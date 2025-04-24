package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/hasansino/goapp/internal/api"
	"github.com/hasansino/goapp/internal/example"
)

// Handler provider for fiber framework
type Handler struct {
	service *example.Service
}

// New provides handlers for its http endpoints
func New(s *example.Service) *Handler {
	return &Handler{service: s}
}

// Register endpoints in fiber framework
func (h *Handler) Register(e *echo.Group) {
	e.GET("/fruits", h.fruits)
	e.GET("/fruits/:id", h.fruitByID)
}

func (h *Handler) fruits(ctx echo.Context) error {
	r, err := h.service.Fruits(ctx.Request().Context())
	if err != nil {
		return h.processError(ctx, err)
	}
	return api.SendJSON(ctx, r)
}

func (h *Handler) fruitByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return api.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	r, err := h.service.FruitByID(ctx.Request().Context(), id)
	if err != nil {
		return h.processError(ctx, err)
	}
	return api.SendJSON(ctx, r)
}
