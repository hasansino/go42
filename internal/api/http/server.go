package http

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	customMiddleware "github.com/hasansino/go42/internal/api/http/middleware"
	"github.com/hasansino/go42/internal/metrics"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type adapterAccessor interface {
	Register(*echo.Group)
}

type rateLimiterAccessor interface {
	Limit(key any) bool
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

	readyStatus atomic.Bool
	rateLimiter rateLimiterAccessor

	tracingEnabled   bool
	swaggerDarkStyle bool

	bodyLimit string
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
	s.e.HTTPErrorHandler = func(err error, ctx echo.Context) {
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
				slog.String("method", ctx.Request().Method),
				slog.String("uri", ctx.Request().RequestURI),
				slog.String("who", "echo.HTTPErrorHandler"),
			}
			if len(panicStack) > 0 {
				slogAttrs = append(slogAttrs, slog.String("stack", string(panicStack)))
			}
			s.l.ErrorContext(ctx.Request().Context(), logMessage, slogAttrs...)
		}

		// if response is not committed, something unexpected happened
		if ctx.Response().Committed {
			return
		}

		if err := SendJSONError(ctx, httpStatus, httpMessage); err != nil {
			s.l.ErrorContext(
				ctx.Request().Context(),
				"failed to send json error response", slog.Any("error", err))
		}
	}

	// panics are handled and passed to the HTTPErrorHandler
	// this middleware should be always the first one in the chain
	s.e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(ctx echo.Context, err error, stack []byte) error {
			return &PanicError{BaseErr: err, Stack: stack}
		},
	}))

	s.e.Use(customMiddleware.NewRateLimiter(s.rateLimiter))

	s.e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
		Skipper: customMiddleware.DefaultSkipper,
		Limit:   "10M",
	}))

	if s.tracingEnabled {
		s.e.Use(otelecho.Middleware(
			"http-server",
			otelecho.WithSkipper(customMiddleware.DefaultSkipper),
		))
	}

	// normal operation logging (http 100-499)
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

	s.e.Use(customMiddleware.NewMetricsCollector())
	s.e.Use(customMiddleware.NewRequestID())

	s.root = s.e.Group("")
	s.root.Static("/static", s.staticRoot)

	s.root.GET("/health", s.health)
	s.root.GET("/ready", s.ready)
	s.readyStatus.Store(true)

	{
		s.v1 = s.e.Group("/api/v1")

		// serve openapi specification files
		s.v1.Static("", s.swaggerRoot+"/v1")

		// embed swagger html template itself
		tmpl := template.Must(template.New("swagger").Parse(swaggerTemplate))
		s.v1.GET("/", func(c echo.Context) error {
			return tmpl.Execute(c.Response(), swaggerTemplateData{
				SpecURLs:  s.parseSpecDir(s.swaggerRoot+"/v1", "/api/v1/"),
				CDN:       swaggerCDNjsdelivr,
				Version:   swaggerUIVersion,
				DarkTheme: true,
			})
		})
	}

	return s
}

func (s *Server) Start(addr string) error {
	return s.e.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

// Register adapters for /
func (s *Server) Register(adapters ...adapterAccessor) {
	for _, p := range adapters {
		p.Register(s.root)
	}
}

// RegisterV1 adapters for /api/v1/
func (s *Server) RegisterV1(adapters ...adapterAccessor) {
	for _, p := range adapters {
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

// parseSpecDir Reads the directory with OpenAPI spec files and returns a map.
// Ignores `.combined.yaml`"` file, which should be generated.
func (s *Server) parseSpecDir(dir string, prefix string) map[string]string {
	specURLs := make(map[string]string)
	specDir, err := os.ReadDir(dir)
	if err != nil {
		s.l.Error(
			"failed to read spec directory",
			slog.String("dir", dir),
			slog.Any("error", err),
		)
		return specURLs
	}
	for _, file := range specDir {
		// ignore .combined.yaml - it should be generated with `make generate`
		if file.IsDir() || file.Name() == ".combined.yaml" {
			continue
		}
		parts := strings.Split(file.Name(), ".")
		if len(parts) != 2 {
			s.l.Warn(
				"unexpected spec file name format",
				slog.String("file", file.Name()),
			)
			continue
		}
		specURLs[parts[0]] = prefix + file.Name()
	}
	return specURLs
}
