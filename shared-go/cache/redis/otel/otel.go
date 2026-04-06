package otel

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"cyber-ecosystem/shared-go/cache"
)

// TracingHook implements cache.Tracer for redis backend.
type TracingHook struct {
	tracer trace.Tracer
}

func NewTracingHook() *TracingHook {
	return &TracingHook{
		tracer: otel.GetTracerProvider().Tracer("cache"),
	}
}

func (h *TracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *TracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		op := strings.ToLower(cmd.Name())
		attrs := []attribute.KeyValue{
			attribute.String("db.system", "cache"),
			attribute.String("db.cache.backend", "redis"),
			attribute.String("db.cache.operation", op),
			attribute.String("db.cache.command", cmd.String()),
			attribute.String("cache.backend", "redis"),
			attribute.String("cache.op", op),
		}
		if keyCount := estimateRedisKeyCount(op, cmd.Args()); keyCount > 0 {
			attrs = append(attrs, attribute.Int("cache.key_count", keyCount))
		}

		newCtx, span := h.tracer.Start(ctx, "cache "+op,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)
		defer span.End()

		start := time.Now()
		err := next(newCtx, cmd)
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
}

func (h *TracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		newCtx, span := h.tracer.Start(ctx, "cache pipeline",
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "cache"),
				attribute.String("db.cache.backend", "redis"),
				attribute.Int("db.cache.pipeline.length", len(cmds)),
				attribute.String("cache.backend", "redis"),
				attribute.String("cache.op", "pipeline"),
				attribute.Int("cache.key_count", len(cmds)),
			),
		)
		defer span.End()

		start := time.Now()
		err := next(newCtx, cmds)
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
}

// TraceOperation implements cache.Tracer (no-op for redis hook pattern)
func (h *TracingHook) TraceOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	return fn(ctx)
}

// TraceOperationWithArgs implements cache.Tracer (no-op for redis hook pattern)
func (h *TracingHook) TraceOperationWithArgs(ctx context.Context, operation string, args *cache.OperationArgs, fn func(context.Context) error) error {
	return fn(ctx)
}

// TracePipeline implements cache.Tracer (no-op for redis hook pattern)
func (h *TracingHook) TracePipeline(ctx context.Context, operations []string, fn func(context.Context) error) error {
	return fn(ctx)
}

func estimateRedisKeyCount(op string, args []interface{}) int {
	argc := len(args)
	if argc <= 1 {
		return 0
	}
	switch op {
	case "mget", "del", "exists", "zrem":
		return argc - 1
	case "mset":
		return (argc - 1) / 2
	default:
		return 1
	}
}

func traceErrorKind(err error) string {
	switch {
	case errors.Is(err, cache.ErrInvalidArgument):
		return "invalid_argument"
	case errors.Is(err, cache.ErrCacheMiss),
		errors.Is(err, cache.ErrKeyNotFound),
		errors.Is(err, cache.ErrSessionNotFound),
		errors.Is(err, cache.ErrQuotaExceeded),
		errors.Is(err, redis.Nil):
		return "expected"
	default:
		return "backend"
	}
}
