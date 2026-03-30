package main

import (
	"context"
	"flag"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	sourcesdk "go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	zaplog "cyber-ecosystem/shared-go/kratos/logging/zap"
	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/conf"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/server"
)

var (
	Name     string = "app_1_service_1"
	Version  string
	flagConf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path, eg: -conf config.yaml")

	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,  // Zero values are emitted
		UseProtoNames:   false, // camelCase output (createdAt, not created_at)
	}
	json.UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignore unknown fields
	}
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cs *connect.Server, os *server.OpsServer) *kratos.App {
	var srv []transport.Server
	if gs != nil {
		srv = append(srv, gs)
	}
	if hs != nil {
		srv = append(srv, hs)
	}
	if cs != nil {
		srv = append(srv, cs)
	}
	if os != nil {
		srv = append(srv, os)
	}
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(srv...),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger, loggerCleanup, err := zaplog.NewLoggerFromConfig(convertLogConfig(bc.Log))
	if err != nil {
		panic(err)
	}
	log.SetLogger(logger)
	logger = log.With(logger,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	meterProvider, metricRequests, metricSeconds, metricsCleanup := setupMetrics()

	var tp *tracesdk.TracerProvider
	if bc.Trace != nil && bc.Trace.Endpoint != "" {
		traceOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(bc.Trace.Endpoint),
		}
		if bc.Trace.Insecure {
			traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(context.Background(), traceOpts...)
		if err != nil {
			panic(err)
		}
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(exp),
			tracesdk.WithResource(sourcesdk.NewSchemaless(semconv.ServiceNameKey.String(Name))),
		)
		otel.SetTracerProvider(tp)
	}

	app, cleanup, err := wireApp(
		bc.Server,
		bc.Auth,
		bc.Log,
		bc.Data,
		bc.Ops,
		logger,
		tp,
		meterProvider,
		metricRequests,
		metricSeconds,
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		cleanup()
		loggerCleanup()
		metricsCleanup()
		if tp != nil {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Errorf("failed to shutdown tracer provider: %v", err)
			}
		}
	}()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func setupMetrics() (*metricsdk.MeterProvider, metric.Int64Counter, metric.Float64Histogram, func()) {
	var safeRegister = func(c prometheus.Collector) {
		if err := prometheus.Register(c); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				return
			}
			panic(err)
		}
	}

	safeRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	safeRegister(collectors.NewGoCollector())

	exporter, err := otelprometheus.New(otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
	if err != nil {
		panic(err)
	}

	mp := metricsdk.NewMeterProvider(
		metricsdk.WithReader(exporter),
		metricsdk.WithResource(sourcesdk.NewSchemaless(semconv.ServiceNameKey.String(Name))),
	)

	meter := mp.Meter(Name)
	metricRequests, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		panic(err)
	}
	metricSeconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		if mp != nil {
			if err := mp.Shutdown(context.Background()); err != nil {
				log.Errorf("failed to shutdown meter provider: %v", err)
			}
		}
	}

	return mp, metricRequests, metricSeconds, cleanup
}

func convertLogConfig(c *conf.Log) *zaplog.Config {
	if c == nil {
		return nil
	}

	cfg := &zaplog.Config{
		Level: c.Level,
	}

	if c.Console != nil {
		cfg.Console = &zaplog.ConsoleConfig{
			Enabled: c.Console.Enabled,
			Color:   c.Console.Color,
			Format:  c.Console.Format,
		}
	}

	if c.File != nil {
		cfg.File = &zaplog.FileConfig{
			Enabled:    c.File.Enabled,
			Path:       c.File.Path,
			MaxSize:    int(c.File.MaxSize),
			MaxBackups: int(c.File.MaxBackups),
			MaxAge:     int(c.File.MaxAge),
			Compress:   c.File.Compress,
		}
	}

	if c.Loki != nil {
		batchWait := int64(1000) // default 1s in milliseconds
		if c.Loki.BatchWait != nil {
			batchWait = c.Loki.BatchWait.AsDuration().Milliseconds()
		}
		cfg.Loki = &zaplog.LokiConfig{
			Enabled:   c.Loki.Enabled,
			URL:       c.Loki.Url,
			Labels:    c.Loki.Labels,
			BatchWait: batchWait,
			BatchSize: int(c.Loki.BatchSize),
		}
	}

	return cfg
}
