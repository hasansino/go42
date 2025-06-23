package tools

import "context"

const ContextKeyRequestID contextKey = "request_id"

type contextKey string

func GetRequestIDFromContext(ctx context.Context) string {
	val, ok := ctx.Value(ContextKeyRequestID).(string)
	if ok {
		return val
	}
	return ""
}

func SetRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}
