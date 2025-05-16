package middleware

import (
	"net/http"
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

			metrics.Counter("application_requests_total", labels).Inc()

			resRecorder := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				status:         200,
			}
			c.Response().Writer = resRecorder

			// --- BEFORE

			err := next(c) // -- APP

			// --- AFTER

			latency := time.Since(start).Seconds()

			responseLabels := map[string]interface{}{
				"method":   labels["method"],
				"path":     labels["path"],
				"status":   strconv.Itoa(resRecorder.status),
				"is_error": toStringBool(err != nil),
			}

			size := float64(resRecorder.size)

			metrics.Counter("application_responses_total", responseLabels).Inc()
			metrics.Histogram("application_request_latency_sec", responseLabels).Update(latency)
			metrics.Histogram("application_response_size_bytes", responseLabels).Update(size)

			return err
		}
	}
}

func toStringBool(is bool) string {
	if is {
		return "yes"
	}
	return "no"
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}
