package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
			slog.String("service", "api"),
			slog.String("who", "echo.StdLogger"),
		).Handler(),
		slog.LevelError,
	)

	// can be used my some middleware, but should be avoided
	e.Logger.SetOutput(slog.NewLogLogger(
		slog.Default().With(
			slog.String("service", "api"),
			slog.String("who", "echo.Logger"),
		).Handler(),
		slog.LevelError,
	).Writer())

	// any (non-panic) errors coming from router eventually land here
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		var echoErr *echo.HTTPError
		if errors.As(err, &echoErr) {
			message = echoErr.Message.(string)
			code = echoErr.Code
		}

		if code >= http.StatusInternalServerError {
			slog.Error("api error",
				slog.Any("error", err),
				slog.String("message", message),
				slog.Int("code", code),
				slog.String("path", c.Path()),
				slog.String("method", c.Request().Method),
				slog.String("who", "echo.HTTPErrorHandler"),
			)
		}

		if c.Response().Committed {
			return
		}

		if err := SendJSONError(c, code, message); err != nil {
			slog.Error("failed to send json error response", slog.Any("error", err))
		}
	}

	// panics are handled here, and NOT passed to the HTTPErrorHandler
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			slog.Error("api panic",
				slog.Any("error", err),
				slog.String("stack", string(stack)),
				slog.String("path", c.Path()),
				slog.String("method", c.Request().Method),
				slog.String("who", "echo.Recover"),
			)
			return nil
		},
	}))

	// Normal operation logging, http 100-499
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.Default().LogAttrs(
					context.Background(), slog.LevelInfo, "request",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("who", "echo.RequestLogger"),
				)
			} else {
				slog.Error("request error",
					slog.Any("error", v.Error),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("who", "echo.RequestLogger"),
				)
			}
			return nil
		},
	}))

	{ // --- @TODO -> Write library to handle metric creation & re-use
		rpsCounter := promauto.NewCounterVec(
			prometheus.CounterOpts{Name: "application_requests_total"},
			[]string{
				"method", "path",
			},
		)
		e.Use(func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				rpsCounter.With(prometheus.Labels{
					"method": c.Request().Method,
					"path":   c.Path(),
				}).Inc()
				return handlerFunc(c)
			}
		})
	} // ---

	for _, opt := range opts {
		opt(e)
	}

	root := e.Group("")

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
	return s.e.Close()
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
