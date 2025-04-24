package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler provider for echo framework
type Handler struct {
	handler http.Handler
}

// New provides handlers for its http endpoints
func New(h http.Handler) *Handler {
	return &Handler{handler: h}
}

// Register endpoints in echo framework
func (h *Handler) Register(e *echo.Group) {
	e.GET("/metrics", echo.WrapHandler(h.handler))
}
