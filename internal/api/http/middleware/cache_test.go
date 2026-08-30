package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	httpMiddleware "github.com/hasansino/go42/internal/api/http/middleware"
	"github.com/hasansino/go42/internal/cache/bigcache"
)

func TestCacheMiddleware_DoesNotCacheNonOKResponses(t *testing.T) {
	cache, err := bigcache.New(bigcache.WithVerbose(false))
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	t.Cleanup(func() {
		if err := cache.Shutdown(context.Background()); err != nil {
			t.Errorf("shut down cache: %v", err)
		}
	})

	var handlerCalls atomic.Int32
	e := echo.New()
	e.GET("/resource", func(c *echo.Context) error {
		handlerCalls.Add(1)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}, httpMiddleware.CacheMiddleware(cache, time.Minute))

	for requestNumber := 1; requestNumber <= 2; requestNumber++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/resource", nil)

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf(
				"request %d: status = %d, want %d",
				requestNumber,
				recorder.Code,
				http.StatusNotFound,
			)
		}
	}

	if got := handlerCalls.Load(); got != 2 {
		t.Errorf("handler calls = %d, want 2; non-OK responses must not be cached", got)
	}
}
