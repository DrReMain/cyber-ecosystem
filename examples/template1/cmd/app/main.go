package main

import (
	"context"
	"flag"
	"os"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/conf"

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
)

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path, eg: -conf config.yaml")
	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	{
		exporter, err := otelprometheus.New(otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
		if err != nil {
			panic(err)
		}
		provider := metricsdk.NewMeterProvider(metricsdk.WithReader(exporter))
		meter := provider.Meter(Name)
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

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
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

	var tp *tracesdk.TracerProvider
	if bc.Trace.Endpoint != "" {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(bc.Trace.Endpoint)))
		if err != nil {
			panic(err)
		}
		tp = tracesdk.NewTracerProvider(tracesdk.WithBatcher(exp), tracesdk.WithResource(sourcesdk.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
		)))
	}

	app, cleanup, err := wireApp(
		bc.Server,
		bc.Data,
		logger,
		tp,
		bc.Metrics,
		_metricRequests,
		_metricSeconds,
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		cleanup()
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
