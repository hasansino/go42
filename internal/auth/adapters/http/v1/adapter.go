package adapter

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/api/http/middleware"
	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	authMiddleware "github.com/hasansino/go42/internal/auth/middleware"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/tools"
)

type serviceAccessor interface {
	SignUp(ctx context.Context, email string, password string) (*models.User, error)
	Login(ctx context.Context, email string, password string) (*domain.Tokens, error)
	Refresh(ctx context.Context, token string) (*domain.Tokens, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	ValidateJWTToken(ctx context.Context, token string) (*jwt.RegisteredClaims, error)
	InvalidateJWTToken(ctx context.Context, token string, until time.Time) error
	ValidateAPIToken(ctx context.Context, token string) (*models.Token, error)
}

type cache interface {
	Get(ctx context.Context, key string) (string, error)
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

type Adapter struct {
	service  serviceAccessor
	cache    cache
	cacheTTL time.Duration
}

func New(service serviceAccessor, opts ...Option) *Adapter {
	a := &Adapter{
		service: service,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *Adapter) Register(g *echo.Group) {
	var cacheMiddleware echo.MiddlewareFunc
	if a.cache != nil && a.cacheTTL > 0 {
		cacheMiddleware = middleware.CacheMiddleware(a.cache, a.cacheTTL)
	}

	authGroup := g.Group("/auth")
	authGroup.POST("/signup", a.signup)
	authGroup.POST("/login", a.login)
	authGroup.POST("/refresh", a.refresh)
	authGroup.POST("/logout", a.logout)

	userGroup := g.Group("/users", authMiddleware.NewAuthMiddleware(a.service))
	userGroup.GET("/me", a.currentUser,
		authMiddleware.NewAccessMiddleware(domain.RBACPermissionUserReadSelf), cacheMiddleware,
	)
}

type SignupRequest struct {
	Email    string `json:"email"    v:"required,email"`
	Password string `json:"password" v:"required,min=8,max=24"`
}

func (a *Adapter) signup(ctx echo.Context) error {
	req := new(SignupRequest)

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

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user, err := a.service.SignUp(ctx.Request().Context(), req.Email, req.Password)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, UserResponseFromModel(user))
}

type LoginRequest struct {
	Email    string `json:"email"    v:"required,email"`
	Password string `json:"password" v:"required,min=8,max=24"`
}

func (a *Adapter) login(ctx echo.Context) error {
	req := new(LoginRequest)

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

	tokens, err := a.service.Login(ctx.Request().Context(), req.Email, req.Password)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, tokens)
}

type RefreshTokenRequest struct {
	Token string `json:"token" v:"required"`
}

func (a *Adapter) refresh(ctx echo.Context) error {
	req := new(RefreshTokenRequest)

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

	tokens, err := a.service.Refresh(ctx.Request().Context(), req.Token)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, tokens)
}

type LogoutTokenRequest struct {
	AccessToken  string `json:"access_token"  v:"required"`
	RefreshToken string `json:"refresh_token" v:"required"`
}

func (a *Adapter) logout(ctx echo.Context) error {
	req := new(LogoutTokenRequest)

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

	err := a.service.Logout(ctx.Request().Context(), req.AccessToken, req.RefreshToken)
	if err != nil {
		return a.processError(ctx, err)
	}
	return ctx.NoContent(http.StatusOK)
}

func (a *Adapter) currentUser(ctx echo.Context) error {
	authInfo := auth.RetrieveAuthFromContext(ctx.Request().Context())
	if authInfo == nil {
		return httpAPI.SendJSONError(ctx,
			http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}

	user, err := a.service.GetUserByID(ctx.Request().Context(), authInfo.ID)
	if err != nil {
		return a.processError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, UserResponseFromModel(user))
}
