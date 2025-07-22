package tools

import (
	"context"
)

const contextKeyRequestID contextKey = "request_id"

type contextKey string

func GetRequestIDFromContext(ctx context.Context) string {
	val, ok := ctx.Value(contextKeyRequestID).(string)
	if ok {
		return val
	}
	return ""
}

func SetRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, requestID)
}
