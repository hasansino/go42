package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/hasansino/goapp/internal/metrics"
)

const (
	ShutdownTimeout = 10 * time.Second
)

type PanicError struct {
	BaseErr error
	Stack   []byte
}

func (e *PanicError) Error() string {
	return e.BaseErr.Error()
}

type Server struct {
	e    *echo.Echo
	root *echo.Group
	v1   *echo.Group
}

func New(opts ...Option) *Server {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	// goes to http.Server.ErrorLog
	// logs low-level errors, like connection or tls errors
	e.StdLogger = slog.NewLogLogger(
		slog.Default().With(
			slog.String("system", "api"),
			slog.String("who", "echo.StdLogger"),
		).Handler(),
		slog.LevelError,
	)

	// can be used my some middleware, but should be avoided
	e.Logger.SetOutput(slog.NewLogLogger(
		slog.Default().With(
			slog.String("system", "api"),
			slog.String("who", "echo.Logger"),
		).Handler(),
		slog.LevelError,
	).Writer())

	// panics are handled and passed to the HTTPErrorHandler
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			return &PanicError{BaseErr: err, Stack: stack}
		},
	}))

	// all panics and explicit errors are handled here
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var (
			httpStatus  = http.StatusInternalServerError
			httpMessage = "Internal Server Error"
		)
		var (
			logMessage = "api error"
			panicStack []byte
		)

		if panicErr := new(PanicError); errors.As(err, &panicErr) {
			logMessage = "api panic"
			panicStack = panicErr.Stack
		} else if echoErr := new(echo.HTTPError); errors.As(err, &echoErr) {
			httpStatus = echoErr.Code
			httpMessage = echoErr.Message.(string)
		}

		slogAttrs := []interface{}{
			slog.String("system", "api"),
			slog.String("error", err.Error()),
			slog.Int("status", httpStatus),
			slog.String("method", c.Request().Method),
			slog.String("uri", c.Request().RequestURI),
			slog.String("who", "echo.HTTPErrorHandler"),
		}
		if len(panicStack) > 0 {
			slogAttrs = append(slogAttrs, slog.String("stack", string(panicStack)))
		}
		slog.Error(logMessage, slogAttrs...)

		// during normal operation, for 4xx errors, response will be already written
		if c.Response().Committed {
			return
		}

		if err := SendJSONError(c, httpStatus, httpMessage); err != nil {
			slog.Error("failed to send json error response", slog.Any("error", err))
		}
	}

	// normal operation logging, http 100-499
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogError:  true,
		LogStatus: true,
		LogMethod: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.InfoContext(
					c.Request().Context(), "request",
					slog.String("system", "api"),
					slog.Int("status", v.Status),
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.String("who", "echo.RequestLogger"),
				)
			}
			return nil
		},
	}))

	// metrics
	e.Use(func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			metrics.Counter("application_requests_total",
				map[string]interface{}{
					"method": c.Request().Method,
				}).Inc()
			return handlerFunc(c)
		}
	})

	for _, opt := range opts {
		opt(e)
	}

	root := e.Group("")
	root.Static("/", "/usr/share/www")

	apiV1 := e.Group("/api/v1")
	apiV1.Static("/", "/usr/share/www/api/v1")
	apiV1.GET("/health-check", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return &Server{e: e, root: root, v1: apiV1}
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.e.Shutdown(ctx)
}

// Register providers for /
func (s *Server) Register(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.root)
	}
}

// RegisterV1 providers for /api/v1/
func (s *Server) RegisterV1(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.v1)
	}
}
