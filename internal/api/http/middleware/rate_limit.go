package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type rateLimiterAcessor interface {
	Limit(key any) bool
}

func NewRateLimiter(limiter rateLimiterAcessor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if limiter == nil {
				return next(c)
			}
			if DefaultSkipper(c) {
				return next(c)
			}
			if !limiter.Limit(extractRateLimitKeyFromCtx(c)) {
				return c.NoContent(http.StatusTooManyRequests)
			}
			return next(c)
		}
	}
}

func extractRateLimitKeyFromCtx(c echo.Context) string {
	return c.RealIP()
}
