package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/hasansino/go42/internal/metrics"
)

func NewMetricsCollector() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {

			start := time.Now()

			labels := map[string]interface{}{
				"method": c.Request().Method,
				"path":   c.Path(),
			}

			metrics.Counter("application_http_requests_count", labels).Inc()

			resRecorder := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				status:         200,
			}
			c.Response().Writer = resRecorder

			// --- BEFORE

			err := next(c) // -- APP

			// --- AFTER

			duration := time.Since(start).Seconds()

			labels["status"] = strconv.Itoa(resRecorder.status)
			labels["is_error"] = "no"
			if err != nil {
				labels["is_error"] = "yes"
			}

			metrics.Counter("application_http_responses_count", labels).Inc()
			metrics.Histogram("application_http_latency_sec", labels).Update(duration)

			return err
		}
	}
}
