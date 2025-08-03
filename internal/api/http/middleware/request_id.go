package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/hasansino/go42/internal/tools"
)

const headerNameRequestID = "x-request-id"

func NewRequestID() echo.MiddlewareFunc {
	return echoMiddleware.RequestIDWithConfig(echoMiddleware.RequestIDConfig{
		Skipper: DefaultSkipper,
		Generator: func() string {
			return uuid.New().String()
		},
		TargetHeader: headerNameRequestID,
		RequestIDHandler: func(ctx echo.Context, requestID string) {
			newCtx := tools.SetRequestIDToContext(ctx.Request().Context(), requestID)
			ctx.SetRequest(ctx.Request().WithContext(newCtx))
			if span := trace.SpanFromContext(newCtx); span.SpanContext().IsValid() && span.IsRecording() {
				span.SetAttributes(attribute.String("http.request_id", requestID))
			}
		},
	})
}
