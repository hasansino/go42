package tools

import (
	"context"
	"log/slog"
)

// ContextFieldsHandler is wrapper around slog handler which automatically adds
// predefined fields to logs taking values from context.
type ContextFieldsHandler struct {
	slog.Handler
}

func SlogContextWrapper(h slog.Handler) slog.Handler {
	return &ContextFieldsHandler{h}
}

func (h *ContextFieldsHandler) Handle(ctx context.Context, r slog.Record) error {
	requestID := GetRequestIDFromContext(ctx)
	if requestID != "" {
		r.Add(slog.String("request-id", GetRequestIDFromContext(ctx)))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *ContextFieldsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextFieldsHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h *ContextFieldsHandler) WithGroup(name string) slog.Handler {
	return &ContextFieldsHandler{
		Handler: h.Handler.WithGroup(name),
	}
}
