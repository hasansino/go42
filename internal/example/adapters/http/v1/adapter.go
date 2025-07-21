package adapter

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/api/http/middleware"
	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
	"github.com/hasansino/go42/internal/tools"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	Fruits(ctx context.Context, limit int, offset int) ([]*models.Fruit, error)
	FruitByID(ctx context.Context, id int) (*models.Fruit, error)
	Create(ctx context.Context, name string) (*models.Fruit, error)
	Update(ctx context.Context, id int, name string) (*models.Fruit, error)
	Delete(ctx context.Context, id int) error
}

type cache interface {
	Get(ctx context.Context, key string) (string, error)
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

// Adapter for fiber framework
type Adapter struct {
	service  serviceAccessor
	cache    cache
	cacheTTL time.Duration
}

// New provides handlers for its http endpoints
func New(s serviceAccessor, c cache, opts ...Option) *Adapter {
	p := &Adapter{service: s, cache: c}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Register endpoints in fiber framework
func (a *Adapter) Register(e *echo.Group) {
	var cacheMiddleware echo.MiddlewareFunc
	if a.cache != nil && a.cacheTTL > 0 {
		cacheMiddleware = middleware.CacheMiddleware(a.cache, a.cacheTTL)
	}
	e.GET("/fruits", a.fruits, cacheMiddleware)
	e.GET("/fruits/:id", a.fruitByID, cacheMiddleware)
	e.POST("/fruits", a.createFruit)
	e.PUT("/fruits/:id", a.updateFruit)
	e.DELETE("/fruits/:id", a.deleteFruit)
}

func (a *Adapter) fruits(ctx echo.Context) error {
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit < 0 {
		limit = domain.DefaultFetchLimit
	}
	offSet, err := strconv.Atoi(ctx.QueryParam("offset"))
	if err != nil || offSet < 0 {
		offSet = 0
	}
	r, err := a.service.Fruits(ctx.Request().Context(), limit, offSet)
	if err != nil {
		return a.processError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, r)
}

func (a *Adapter) fruitByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	r, err := a.service.FruitByID(ctx.Request().Context(), id)
	if err != nil {
		return a.processError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, r)
}

type CreateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=255"`
}

func (a *Adapter) createFruit(ctx echo.Context) error {
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

	r, err := a.service.Create(ctx.Request().Context(), req.Name)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, r)
}

type UpdateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=255"`
}

func (a *Adapter) updateFruit(ctx echo.Context) error {
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

	r, err := a.service.Update(ctx.Request().Context(), id, req.Name)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, r)
}

func (a *Adapter) deleteFruit(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}
	if err := a.service.Delete(ctx.Request().Context(), id); err != nil {
		return a.processError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}
