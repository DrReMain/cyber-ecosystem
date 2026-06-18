package observability

import (
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/metrics"
	"github.com/go-kratos/kratos/v3/middleware"
	"go.opentelemetry.io/otel"
)

// MetricsServer builds the kratos contrib metrics middleware (server side) with
// the default server instruments (request counter + latency histogram), created
// from the global meter provider set by Init. Call after Init, in the server
// middleware chain. Mirrors contrib metrics.Server().
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
