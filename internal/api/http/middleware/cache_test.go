package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"

	httpMiddleware "github.com/hasansino/go42/internal/api/http/middleware"
	middlewareMocks "github.com/hasansino/go42/internal/api/http/middleware/mocks"
)

func TestCacheMiddleware_DoesNotCacheNonOKResponses(t *testing.T) {
	ctrl := gomock.NewController(t)
	cache := middlewareMocks.NewMockcacheAccessor(ctrl)
	cache.EXPECT().
		Get(gomock.Any(), "GET+/resource?").
		Return("", false, nil).
		Times(2)

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

func TestCacheMiddleware_CachesEmptyOKResponses(t *testing.T) {
	ctrl := gomock.NewController(t)
	cache := middlewareMocks.NewMockcacheAccessor(ctrl)
	gomock.InOrder(
		cache.EXPECT().
			Get(gomock.Any(), "GET+/resource?").
			Return("", false, nil),
		cache.EXPECT().
			Set(gomock.Any(), "GET+/resource?", "", time.Minute).
			Return(nil),
		cache.EXPECT().
			Get(gomock.Any(), "GET+/resource?").
			Return("", true, nil),
	)

	var handlerCalls atomic.Int32
	e := echo.New()
	e.GET("/resource", func(c *echo.Context) error {
		handlerCalls.Add(1)
		return c.NoContent(http.StatusOK)
	}, httpMiddleware.CacheMiddleware(cache, time.Minute))

	expectedCacheHeaders := []string{"MISS", "HIT"}
	for requestNumber, expectedCacheHeader := range expectedCacheHeaders {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/resource", nil)

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf(
				"request %d: status = %d, want %d",
				requestNumber+1,
				recorder.Code,
				http.StatusOK,
			)
		}
		if got := recorder.Header().Get("X-Cache"); got != expectedCacheHeader {
			t.Errorf(
				"request %d: X-Cache = %q, want %q",
				requestNumber+1,
				got,
				expectedCacheHeader,
			)
		}
		if recorder.Body.Len() != 0 {
			t.Errorf("request %d: body = %q, want empty", requestNumber+1, recorder.Body.String())
		}
	}

	if got := handlerCalls.Load(); got != 1 {
		t.Errorf("handler calls = %d, want 1; empty successful responses must be cached", got)
	}
}
