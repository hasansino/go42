package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/goapp/internal/api/http"
	"github.com/hasansino/goapp/internal/example"
	"github.com/hasansino/goapp/internal/example/domain"
	"github.com/hasansino/goapp/internal/tools"
)

// Provider Provider for fiber framework
type Provider struct {
	service *example.Service
}

// New provides handlers for its http endpoints
func New(s *example.Service) *Provider {
	return &Provider{service: s}
}

// Register endpoints in fiber framework
func (h *Provider) Register(e *echo.Group) {
	e.GET("/fruits", h.fruits)
	e.GET("/fruits/:id", h.fruitByID)
	e.POST("/fruits", h.createFruit)
	e.PUT("/fruits/:id", h.updateFruit)
	e.DELETE("/fruits/:id", h.deleteFruit)
}

func (h *Provider) fruits(ctx echo.Context) error {
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
	return ctx.JSON(http.StatusOK, r)
}

func (h *Provider) fruitByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	r, err := h.service.FruitByID(ctx.Request().Context(), id)
	if err != nil {
		return h.processError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, r)
}

func (h *Provider) createFruit(ctx echo.Context) error {
	req := new(domain.CreateFruitRequest)

	if err := ctx.Bind(req); err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	vErrs := tools.ValidateStruct(req)
	if vErrs != nil {
		return httpAPI.SendJSONError(
			ctx, http.StatusBadRequest, http.StatusText(http.StatusBadRequest),
			httpAPI.WithValidationErrors(vErrs...),
		)
	}

	r, err := h.service.Create(ctx.Request().Context(), req)
	if err != nil {
		return h.processError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, r)
}

func (h *Provider) deleteFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	if err := h.service.Delete(ctx.Request().Context(), id); err != nil {
		return h.processError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}

func (h *Provider) updateFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	req := new(domain.UpdateFruitRequest)

	if err := ctx.Bind(req); err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	vErrs := tools.ValidateStruct(req)
	if vErrs != nil {
		return httpAPI.SendJSONError(
			ctx, http.StatusBadRequest, http.StatusText(http.StatusBadRequest),
			httpAPI.WithValidationErrors(vErrs...),
		)
	}

	r, err := h.service.Update(ctx.Request().Context(), id, req)
	if err != nil {
		return h.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, r)
}
