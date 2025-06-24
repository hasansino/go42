package http

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	customMiddleware "github.com/hasansino/go42/internal/api/http/middleware"
	"github.com/hasansino/go42/internal/metrics"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

// providerAccessor for all handler providers.
type providerAccessor interface {
	Register(*echo.Group)
}

type PanicError struct {
	BaseErr error
	Stack   []byte
}

func (e *PanicError) Error() string {
	return e.BaseErr.Error()
}

type Server struct {
	l    *slog.Logger
	e    *echo.Echo
	root *echo.Group
	v1   *echo.Group

	staticRoot  string
	swaggerRoot string

	tracingEnabled bool
	readyStatus    atomic.Bool
}

func New(opts ...Option) *Server {
	s := &Server{
		e: echo.New(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.e.HideBanner = true
	s.e.HidePort = true

	// goes to http.HTTPServer.ErrorLog
	// logs low-level errors, like connection or tls errors
	s.e.StdLogger = slog.NewLogLogger(
		s.l.Handler().WithAttrs([]slog.Attr{
			slog.String("who", "echo.StdLogger"),
		}),
		slog.LevelError,
	)

	// can be used my some middleware, but should be avoided
	s.e.Logger.SetOutput(slog.NewLogLogger(
		s.l.Handler().WithAttrs([]slog.Attr{
			slog.String("who", "echo.Logger"),
		}),
		slog.LevelError,
	).Writer())

	// all panics and explicit errors are handled here
	s.e.HTTPErrorHandler = func(err error, c echo.Context) {
		var (
			httpStatus  = http.StatusInternalServerError
			httpMessage = "Internal HTTPServer Error"
		)

		var (
			logMessage      = "http api error"
			metricErrorType = "http_api_error"
			panicStack      []byte
		)

		if panicErr := new(PanicError); errors.As(err, &panicErr) {
			logMessage = "http api panic"
			metricErrorType = "http_api_panic"
			panicStack = panicErr.Stack
		} else if echoErr := new(echo.HTTPError); errors.As(err, &echoErr) {
			httpStatus = echoErr.Code
			httpMessage = http.StatusText(httpStatus)
		}

		if httpStatus >= 500 {
			metrics.Counter("errors", map[string]interface{}{
				"type": metricErrorType,
			}).Inc()
			slogAttrs := []interface{}{
				slog.String("error", err.Error()),
				slog.Int("status", httpStatus),
				slog.String("method", c.Request().Method),
				slog.String("uri", c.Request().RequestURI),
				slog.String("who", "echo.HTTPErrorHandler"),
			}
			if len(panicStack) > 0 {
				slogAttrs = append(slogAttrs, slog.String("stack", string(panicStack)))
			}
			s.l.ErrorContext(c.Request().Context(), logMessage, slogAttrs...)
		}

		// if response is not commited, something unexpected happened
		if c.Response().Committed {
			return
		}

		if err := SendJSONError(c, httpStatus, httpMessage); err != nil {
			s.l.ErrorContext(
				c.Request().Context(),
				"failed to send json error response", slog.Any("error", err))
		}
	}

	// 1. panics are handled and passed to the HTTPErrorHandler
	s.e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			return &PanicError{BaseErr: err, Stack: stack}
		},
	}))

	// 2. metric collector
	s.e.Use(customMiddleware.NewMetricsCollector())

	// 3. normal operation logging (http 100-499)
	s.e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:      customMiddleware.DefaultSkipper,
		LogError:     true,
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// log any request with status code < 500 as normal INFO level
			echoErr := new(echo.HTTPError)
			if v.Error == nil || errors.As(v.Error, &echoErr) && echoErr.Code < 500 {
				s.l.DebugContext(
					c.Request().Context(),
					"request",
					slog.Int("status", v.Status),
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.String("who", "echo.RequestLogger"),
				)
			}
			return nil
		},
	}))

	// 4. tracing
	if s.tracingEnabled {
		s.e.Use(otelecho.Middleware("http-server"))
	}

	// 5. request id
	s.e.Use(customMiddleware.NewRequestID())

	root := s.e.Group("")
	root.Static("/", s.staticRoot)

	root.GET("/health", s.health)
	root.GET("/ready", s.ready)
	s.readyStatus.Store(true)

	apiV1 := s.e.Group("/api/v1")

	// serve openapi specification files
	apiV1.Static("", s.swaggerRoot+"/v1")

	// embed swagger html template itself
	tmpl := template.Must(template.New("swagger").Parse(swaggerTemplate))
	apiV1.GET("/", func(c echo.Context) error {
		return tmpl.Execute(c.Response(), swaggerTemplateData{
			SpecURL: "/api/v1/openapi.yml",
			CDN:     swaggerCDNjsdelivr,
			Version: swaggerUIVersion,
		})
	})

	s.root = root
	s.v1 = apiV1

	return s
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
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

func (s *Server) health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (s *Server) ready(ctx echo.Context) error {
	if !s.readyStatus.Load() {
		return ctx.NoContent(http.StatusServiceUnavailable)
	}
	return ctx.NoContent(http.StatusOK)
}
