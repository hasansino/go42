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

	// panics are handled and passed to the HTTPErrorHandler
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			return &PanicError{BaseErr: err, Stack: stack}
		},
	}))

	// all panics and errors are handled here
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var (
			httpCode    = http.StatusInternalServerError
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
			httpMessage = echoErr.Message.(string)
			httpCode = echoErr.Code
		}

		// not all 5xx are unexpected
		if httpCode >= http.StatusInternalServerError {
			slogAttrs := []interface{}{
				slog.String("error", err.Error()),
				slog.String("httpMessage", httpMessage),
				slog.Int("httpCode", httpCode),
				slog.String("path", c.Path()),
				slog.String("method", c.Request().Method),
				slog.String("who", "echo.HTTPErrorHandler"),
			}
			if len(panicStack) > 0 {
				slogAttrs = append(slogAttrs, slog.String("stack", string(panicStack)))
			}
			slog.Error(logMessage, slogAttrs...)
		}

		// during normal operation, for 4xx errors, response will be already written
		if c.Response().Committed {
			return
		}

		if err := SendJSONError(c, httpCode, httpMessage); err != nil {
			slog.Error("failed to send json error response", slog.Any("error", err))
		}
	}

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
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Any("error", v.Error),
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
