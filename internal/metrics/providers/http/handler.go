package http

import (
	"net/http"

	"github.com/labstack/echo"
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
func (h *Handler) Register(e *echo.Group) {
	e.GET("/metrics", echo.WrapHandler(h.handler))
}
