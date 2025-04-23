package api

import (
	"time"

	"github.com/labstack/echo"
)

type Option func(e *echo.Echo)

func WithReadTimeout(d time.Duration) Option {
	return func(e *echo.Echo) {
		e.Server.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(e *echo.Echo) {
		e.Server.WriteTimeout = d
	}
}
