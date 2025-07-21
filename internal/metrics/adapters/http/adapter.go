package adapter

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Adapter for echo framework
type Adapter struct {
	handler http.Handler
}

// New provides handlers for its http endpoints
func New(h http.Handler) *Adapter {
	return &Adapter{handler: h}
}

// Register endpoints in echo framework
func (h *Adapter) Register(e *echo.Group) {
	e.GET("/metrics", echo.WrapHandler(h.handler))
}
