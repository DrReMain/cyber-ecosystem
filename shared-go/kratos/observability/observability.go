package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/contrib/otel/v3/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	metricSDK "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init configures the global OTel providers (tracer/meter) and returns a
// multi-sink *slog.Logger, all per cfg. The contrib tracing/metrics middlewares
// and the log sinks read these globals. Service identity (name/version/
// instanceID) is passed in so this package stays service-agnostic. The returned
// shutdown flushes every enabled provider (each with its own timeout) and must
// be deferred.
func Init(cfg Config, name, version, instanceID string) (func(), *slog.Logger, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.id", instanceID),
			attribute.String("service.name", name),
			attribute.String("service.version", version),
		),
		resource.WithHost(),    // host.name, host.arch
		resource.WithFromEnv(), // OTEL_RESOURCE_ATTRIBUTES (deployment.environment, ...)
	)
	if err != nil {
		return nil, nil, err
	}

	// A signal enabled with no collector endpoint is almost certainly a config
	// mistake — warn loudly instead of silently exporting nothing.
	if cfg.Endpoint == "" && (cfg.Trace.Enabled || cfg.Metrics.Enabled || cfg.Log.OTLP) {
		fmt.Fprintln(os.Stderr, "observability: warning: trace/metrics/otlp-log enabled but no endpoint configured; those signals will not export")
	}

	var shutdowns []func(context.Context) error

	if cfg.Endpoint != "" {
		if cfg.Trace.Enabled {
			opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
			if cfg.Insecure {
				opts = append(opts, otlptracehttp.WithInsecure())
			}
			exp, err := otlptracehttp.New(context.Background(), opts...)
			if err != nil {
				return nil, nil, err
			}
			ratio := cfg.Trace.SamplingRatio
			if ratio <= 0 || ratio > 1 {
				ratio = 1 // default: no sampling (rely on collector tail-sampling)
			}
			tp := sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				sdktrace.WithBatcher(exp),
				sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
			)
			otel.SetTracerProvider(tp)
			shutdowns = append(shutdowns, tp.Shutdown)
		}

		if cfg.Metrics.Enabled {
			opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(cfg.Endpoint)}
			if cfg.Insecure {
				opts = append(opts, otlpmetrichttp.WithInsecure())
			}
			exp, err := otlpmetrichttp.New(context.Background(), opts...)
			if err != nil {
				return nil, nil, err
			}
			mp := metricSDK.NewMeterProvider(
				metricSDK.WithResource(res),
				metricSDK.WithReader(metricSDK.NewPeriodicReader(exp)),
				metricSDK.WithView(metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName)),
			)
			otel.SetMeterProvider(mp)
			shutdowns = append(shutdowns, mp.Shutdown)
		}
	}

	logger := buildLogger(cfg, name, version, instanceID, res, &shutdowns)

	// Each provider gets its own fresh timeout so one slow flush can't starve
	// the others; failures are surfaced (best-effort) rather than dropped.
	return func() {
		for _, s := range shutdowns {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s(ctx); err != nil {
				fmt.Fprintln(os.Stderr, "observability: warning: provider shutdown failed:", err)
			}
			cancel()
		}
	}, logger, nil
}
