package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
	e.StdLogger = log.Default()
	e.Logger.SetOutput(log.Default().Writer())

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{}))
	e.Use(middleware.RecoverWithConfig(middleware.DefaultRecoverConfig))

	{ // --- @TODO -> Write library to handle metric creation & re-use
		rpsCounter := promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "application_requests_total",
		}, []string{
			"method", "path",
		})

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
