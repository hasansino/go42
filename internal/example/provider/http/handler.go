package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/hasansino/goapp/internal/api"
	"github.com/hasansino/goapp/internal/example"
	"github.com/hasansino/goapp/internal/example/domain"
	"github.com/hasansino/goapp/internal/utils"
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
	e.POST("/fruits", h.createFruit)
	e.PUT("/fruits/:id", h.updateFruit)
	e.DELETE("/fruits/:id", h.deleteFruit)
}

func (h *Handler) fruits(ctx echo.Context) error {
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil {
		limit = domain.DefaultFetchLimit
	}
	offSet, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil {
		offSet = 0
	}
	r, err := h.service.Fruits(ctx.Request().Context(), limit, offSet)
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

func (h *Handler) createFruit(ctx echo.Context) error {
	req := new(domain.CreateFruitRequest)

	if err := ctx.Bind(req); err != nil {
		return api.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	vErrs := utils.ValidateStruct(req)
	if vErrs != nil {
		return api.SendJSONValidationError(ctx, vErrs)
	}

	r, err := h.service.Create(ctx.Request().Context(), req)
	if err != nil {
		return h.processError(ctx, err)
	}

	return api.SendJSON(ctx, r)
}

func (h *Handler) deleteFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return api.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	if err := h.service.Delete(ctx.Request().Context(), id); err != nil {
		return h.processError(ctx, err)
	}
	return api.SendJSON(ctx, http.StatusOK)
}

func (h *Handler) updateFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return api.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	req := new(domain.UpdateFruitRequest)

	if err := ctx.Bind(req); err != nil {
		return api.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	vErrs := utils.ValidateStruct(req)
	if vErrs != nil {
		return api.SendJSONValidationError(ctx, vErrs)
	}

	r, err := h.service.Update(ctx.Request().Context(), id, req)
	if err != nil {
		return h.processError(ctx, err)
	}

	return api.SendJSON(ctx, r)
}
