package tools

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Traced is a helper that wraps a function with tracing
func Traced[T any](
	ctx context.Context,
	tracerName string,
	spanName string,
	fn func(context.Context) (T, error),
) (T, error) {
	tracer := otel.Tracer(tracerName)
	if tracer == nil {
		return fn(ctx)
	}
	tracerCtx, span := tracer.Start(ctx, spanName)
	if span == nil || !span.IsRecording() {
		return fn(ctx)
	}
	defer span.End()
	result, err := fn(tracerCtx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	return result, err
}

// TracedAttrs is a helper that wraps a function with tracing with attributes.
func TracedAttrs[T any](
	ctx context.Context,
	tracerName string,
	spanName string,
	attrs []attribute.KeyValue,
	fn func(context.Context) (T, error),
) (T, error) {
	tracer := otel.Tracer(tracerName)
	if tracer == nil {
		return fn(ctx)
	}
	tracerCtx, span := tracer.Start(ctx, spanName, trace.WithAttributes(attrs...))
	if span == nil || !span.IsRecording() {
		return fn(ctx)
	}
	defer span.End()
	result, err := fn(tracerCtx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	return result, err
}

// TracedNoReturn is a helper for functions that don't return a value.
func TracedNoReturn(
	ctx context.Context,
	tracerName string,
	spanName string,
	fn func(context.Context) error,
) error {
	tracer := otel.Tracer(tracerName)
	if tracer == nil {
		return fn(ctx)
	}
	tracerCtx, span := tracer.Start(ctx, spanName)
	if span == nil || !span.IsRecording() {
		return fn(ctx)
	}
	defer span.End()
	err := fn(tracerCtx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	return err
}

// TracedNoReturnAttrs is a helper for functions that don't return a value, with attributes.
func TracedNoReturnAttrs(
	ctx context.Context,
	tracerName string,
	spanName string,
	attrs []attribute.KeyValue,
	fn func(context.Context) error,
) error {
	tracer := otel.Tracer(tracerName)
	if tracer == nil {
		return fn(ctx)
	}
	tracerCtx, span := tracer.Start(ctx, spanName, trace.WithAttributes(attrs...))
	if span == nil || !span.IsRecording() {
		return fn(ctx)
	}
	defer span.End()
	err := fn(tracerCtx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	return err
}
