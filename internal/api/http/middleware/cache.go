package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	cacheHeaderName      = "X-Cache"
	cacheHeaderValueHit  = "HIT"
	cacheHeaderValueMISS = "MISS"
)

type cacheAccessor interface {
	Get(ctx context.Context, key string) (string, error)
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

func CacheMiddleware(cache cacheAccessor, ttl time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := fmt.Sprintf(
				"%s+%s?%s",
				c.Request().Method, c.Request().URL.Path, c.Request().URL.RawQuery,
			)

			cached, err := cache.Get(c.Request().Context(), key)
			if err != nil {
				slog.Default().
					With(slog.String("component", "echo-middleware-cache")).
					Error("failed to fetch cached data", slog.Any("error", err))
				return next(c)
			}

			if len(cached) > 0 {
				c.Response().Header().Add(cacheHeaderName, cacheHeaderValueHit)
				return c.JSONBlob(200, []byte(cached))
			} else {
				c.Response().Header().Add(cacheHeaderName, cacheHeaderValueMISS)
			}

			resRecorder := newResponseRecorder(c.Response().Writer, true)
			c.Response().Writer = resRecorder

			if err := next(c); err != nil {
				return err
			}

			err = cache.SetTTL(c.Request().Context(), key, resRecorder.body.String(), ttl)
			if err != nil {
				slog.Default().
					With(slog.String("component", "echo-middleware-cache")).
					Error("failed to cache response", slog.Any("error", err))
			}

			return nil
		}
	}
}
