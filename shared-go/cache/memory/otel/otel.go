package otel

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"cyber-ecosystem/shared-go/cache"
)

// TracingHook implements cache.Tracer for memory backend.
type TracingHook struct {
	tracer trace.Tracer
}

func NewTracingHook() *TracingHook {
	return &TracingHook{
		tracer: otel.GetTracerProvider().Tracer("cache"),
	}
}

func (h *TracingHook) TraceOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	return h.TraceOperationWithArgs(ctx, operation, nil, fn)
}

func (h *TracingHook) TraceOperationWithArgs(ctx context.Context, operation string, args *cache.OperationArgs, fn func(context.Context) error) error {
	op := strings.ToLower(operation)
	spanAttributes := []attribute.KeyValue{
		attribute.String("db.system", "cache"),
		attribute.String("db.cache.backend", "memory"),
		attribute.String("db.cache.operation", op),
		attribute.String("cache.backend", "memory"),
		attribute.String("cache.op", op),
	}

	if args != nil {
		keyCount := 0
		if args.Key != "" {
			spanAttributes = append(spanAttributes, attribute.String("db.cache.key", args.Key))
			keyCount = 1
		}
		if len(args.Keys) > 0 {
			spanAttributes = append(spanAttributes, attribute.StringSlice("db.cache.keys", args.Keys))
			keyCount = len(args.Keys)
		}
		if len(args.Value) > 0 && len(args.Value) <= 256 {
			spanAttributes = append(spanAttributes, attribute.String("db.cache.value", string(args.Value)))
		}
		if args.SessionID != "" {
			spanAttributes = append(spanAttributes, attribute.String("db.cache.session_id", args.SessionID))
		}
		if args.Member != "" {
			spanAttributes = append(spanAttributes, attribute.String("db.cache.member", args.Member))
		}
		if args.Score != 0 {
			spanAttributes = append(spanAttributes, attribute.Float64("db.cache.score", args.Score))
		}
		if args.Delta != 0 {
			spanAttributes = append(spanAttributes, attribute.Float64("db.cache.delta", args.Delta))
		}
		if args.DeltaInt != 0 {
			spanAttributes = append(spanAttributes, attribute.Int64("db.cache.delta_int", args.DeltaInt))
		}
		if args.Quota != 0 {
			spanAttributes = append(spanAttributes, attribute.Int64("db.cache.quota", args.Quota))
		}
		if args.Window != 0 {
			spanAttributes = append(spanAttributes, attribute.Int64("db.cache.window_ms", args.Window))
		}
		if len(args.Values) > 0 {
			keyCount = len(args.Values)
		}
		if len(args.Members) > 0 {
			keyCount = len(args.Members)
		}
		if keyCount > 0 {
			spanAttributes = append(spanAttributes, attribute.Int("cache.key_count", keyCount))
		}
	}

	newCtx, span := h.tracer.Start(ctx, "cache "+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(spanAttributes...),
	)
	defer span.End()

	start := time.Now()
	err := fn(newCtx)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Int64("db.cache.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.kind", traceErrorKind(err)))
	}
	return err
}

func (h *TracingHook) TracePipeline(ctx context.Context, operations []string, fn func(context.Context) error) error {
	newCtx, span := h.tracer.Start(ctx, "cache pipeline",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "cache"),
			attribute.String("db.cache.backend", "memory"),
			attribute.Int("db.cache.pipeline.length", len(operations)),
			attribute.String("cache.backend", "memory"),
			attribute.String("cache.op", "pipeline"),
			attribute.Int("cache.key_count", len(operations)),
		),
	)
	defer span.End()

	err := fn(newCtx)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.kind", traceErrorKind(err)))
	}
	return err
}

func traceErrorKind(err error) string {
	switch {
	case errors.Is(err, cache.ErrInvalidArgument):
		return "invalid_argument"
	case errors.Is(err, cache.ErrCacheMiss),
		errors.Is(err, cache.ErrKeyNotFound),
		errors.Is(err, cache.ErrSessionNotFound),
		errors.Is(err, cache.ErrQuotaExceeded):
		return "expected"
	default:
		return "backend"
	}
}
