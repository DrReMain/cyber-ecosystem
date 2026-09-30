package observability

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/metrics"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// MetricsServer builds the kratos contrib metrics middleware (server side) with
// the default server instruments (request counter + latency histogram), created
// from the global meter provider set by Init. Call after Init, in the server
// middleware chain. The instruments are supplied here because a bare contrib
// metrics.Server() without options creates none and records nothing.
//
// No MetricsClient: server-side metrics cover the callee view; client-side
// (outbound) metrics are an optional dependency-monitoring nicety and the
// downstream's own server metrics already record those calls. Add a client
// variant only if caller-side outbound metrics are explicitly wanted.
func MetricsServer() middleware.Middleware {
	meter := otel.Meter("kratos")
	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		slog.Warn("observability: metrics counter creation failed", "error", err)
	}
	histogram, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		slog.Warn("observability: metrics histogram creation failed", "error", err)
	}
	return metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram))
}

// ErrorTelemetryServer makes failures first-class telemetry (server side,
// matching the kratos Server/Client middleware naming convention): span attributes with
// the flattened classification (error.reason / error.internal — the
// aggregation keys), an error counter with a reason dimension, and an outbound
// re-rendering of the error so every downstream consumer of Error() — the
// stock tracing middleware's Exception event and status description above all
// — sees the flattened view rather than the recursive chain dump. The
// Exception event itself stays with the stock tracing middleware: recording a
// second event here would split every failure into two error groups
// downstream. ErrorTelemetryServer observes the raw pre-mask error, so it must
// sit inside sanitize in the chain. Like MetricsServer, call after Init.
//
// No client variant exists yet: outbound telemetry lands together with the
// first real remote RP client (client.go), where the same three duties apply
// against the pre-map upstream error — attrs, a kratos.client.errors counter,
// and the outbound re-rendering for the stock client tracing middleware.
func ErrorTelemetryServer() middleware.Middleware {
	counter, err := otel.Meter("kratos").Int64Counter("kratos.server.errors")
	if err != nil {
		slog.Warn("observability: error counter creation failed", "error", err)
		counter = nil
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err == nil {
				return reply, nil
			}
			v := viewError(err)
			// Non-kratos failures (a bug somewhere on the path) still count;
			// they are normalized on the wire by sanitize but keep no identity
			// of their own here.
			if v.code == 0 {
				v.code = 500
			}
			if v.reason == "" {
				v.reason = "unclassified"
			}
			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				attrs := []attribute.KeyValue{
					attribute.String("error.reason", v.reason),
				}
				if path := v.internalPath(); path != "" {
					attrs = append(attrs, attribute.String("error.internal", path))
				}
				span.SetAttributes(attrs...)
			}
			if counter != nil {
				counter.Add(ctx, 1, metric.WithAttributes(
					attribute.String("operation", operationFrom(ctx)),
					attribute.String("reason", v.reason),
					attribute.Int("code", v.code),
				))
			}
			// The outbound error is re-rendered, not replaced: identity
			// queries (As/Unwrap/FromError) still resolve to the original
			// chain, while the stock tracing middleware's RecordError and
			// SetStatus — which consume Error() verbatim — emit the flattened
			// view instead of kratos v3's recursive chain dump.
			return reply, renderFlat(err)
		}
	}
}

// operationFrom returns the RPC operation of the inbound server transport, or
// "" outside one (the counter then aggregates under an empty operation).
func operationFrom(ctx context.Context) string {
	if info, ok := transport.FromServerContext(ctx); ok {
		return info.Operation()
	}
	return ""
}
