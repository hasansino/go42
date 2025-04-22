package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// Handler provider for fiber framework
type Handler struct {
	handler http.Handler
}

// New provides handlers for its http endpoints
func New(h http.Handler) *Handler {
	return &Handler{handler: h}
}

// Register endpoints in fiber framework
func (h *Handler) Register(r fiber.Router) {
	r.Get("/metrics", adaptor.HTTPHandler(h.handler))
}
