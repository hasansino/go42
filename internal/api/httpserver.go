package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	f  *fiber.App
	v1 fiber.Router
}

func New(opts ...Option) *Server {
	cfg := fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			message := "Unexpected API Error"
			code := fiber.StatusInternalServerError
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				message = fiberErr.Message
				code = fiberErr.Code
			}
			slog.ErrorContext(ctx.UserContext(), "API Error", slog.Any("error", err))
			return SendJSONError(ctx, code, message)
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	f := fiber.New(cfg)
	f.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	apiV1 := f.Group("/apiV1/v1")
	apiV1.Static("/", "/usr/share/www/apiV1/v1")
	apiV1.Get("/health-check", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	return &Server{f: f, v1: apiV1}
}

func (s *Server) Start(addr string) error {
	return s.f.Listen(addr)
}

func (s *Server) Close() error {
	return s.f.Shutdown()
}

// Register providers for /
func (s *Server) Register(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.f)
	}
}

// RegisterV1 providers for /api/v1/
func (s *Server) RegisterV1(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.v1)
	}
}
