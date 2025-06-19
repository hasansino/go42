package provider

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/example"
	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/tools"
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
func (p *Provider) Register(e *echo.Group) {
	e.GET("/fruits", p.fruits)
	e.GET("/fruits/:id", p.fruitByID)
	e.POST("/fruits", p.createFruit)
	e.PUT("/fruits/:id", p.updateFruit)
	e.DELETE("/fruits/:id", p.deleteFruit)
}

func (p *Provider) fruits(ctx echo.Context) error {
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit < 0 {
		limit = domain.DefaultFetchLimit
	}
	offSet, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil || offSet < 0 {
		offSet = 0
	}
	r, err := p.service.Fruits(ctx.Request().Context(), limit, offSet)
	if err != nil {
		return p.processError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, r)
}

func (p *Provider) fruitByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	r, err := p.service.FruitByID(ctx.Request().Context(), id)
	if err != nil {
		return p.processError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, r)
}

type CreateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=255"`
}

func (p *Provider) createFruit(ctx echo.Context) error {
	req := new(CreateFruitRequest)

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

	r, err := p.service.Create(ctx.Request().Context(), req.Name)
	if err != nil {
		return p.processError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, r)
}

type UpdateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=255"`
}

func (p *Provider) updateFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	req := new(UpdateFruitRequest)

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

	r, err := p.service.Update(ctx.Request().Context(), id, req.Name)
	if err != nil {
		return p.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, r)
}

func (p *Provider) deleteFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	if err := p.service.Delete(ctx.Request().Context(), id); err != nil {
		return p.processError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}
