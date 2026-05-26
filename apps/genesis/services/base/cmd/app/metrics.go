package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
)

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
		metricsdk.WithResource(newResource()),
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
