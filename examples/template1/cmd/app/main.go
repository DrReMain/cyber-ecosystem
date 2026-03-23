package main

import (
	"context"
	"flag"
	"os"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"

	zaplog "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/logging/zap"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	sourcesdk "go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	Name     string = "examples-template1"
	Version  string
	flagConf string
	id, _    = os.Hostname()

	_metricRequests metric.Int64Counter
	_metricSeconds  metric.Float64Histogram
	_meterProvider  *metricsdk.MeterProvider
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

	{
		exporter, err := otelprometheus.New(otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
		if err != nil {
			panic(err)
		}
		_meterProvider = metricsdk.NewMeterProvider(metricsdk.WithReader(exporter))
		meter := _meterProvider.Meter(Name)
		_metricRequests, err = metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
		if err != nil {
			panic(err)
		}
		_metricSeconds, err = metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
		if err != nil {
			panic(err)
		}
	}
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cs *connect.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
			cs,
		),
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

	var tp *tracesdk.TracerProvider
	if bc.Trace != nil && bc.Trace.Endpoint != "" {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(bc.Trace.Endpoint)))
		if err != nil {
			panic(err)
		}
		tp = tracesdk.NewTracerProvider(tracesdk.WithBatcher(exp), tracesdk.WithResource(sourcesdk.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
		)))
		otel.SetTracerProvider(tp)
	}

	app, cleanup, err := wireApp(
		bc.Server,
		bc.Log,
		bc.Data,
		bc.Metrics,
		logger,
		tp,
		_metricRequests,
		_metricSeconds,
		_meterProvider,
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		cleanup()
		loggerCleanup()
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
