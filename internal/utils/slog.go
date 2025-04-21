package utils

import (
	"context"
	"log/slog"
)

// SlogMultiHandler sends logs to multiple handlers
type SlogMultiHandler struct {
	handlers []slog.Handler
}

// NewSlogMultiHandler creates a new SlogMultiHandler
func NewSlogMultiHandler(handlers ...slog.Handler) *SlogMultiHandler {
	return &SlogMultiHandler{handlers: handlers}
}

// Enabled implements slog.Handler
func (h *SlogMultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle implements slog.Handler
func (h *SlogMultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs implements slog.Handler
func (h *SlogMultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewSlogMultiHandler(handlers...)
}

// WithGroup implements slog.Handler
func (h *SlogMultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewSlogMultiHandler(handlers...)
}
