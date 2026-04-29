package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/template/services/base/internal/conf"
)

func newTracerProvider(bc *conf.Bootstrap) (*tracesdk.TracerProvider, func(), error) {
	if bc.Trace == nil || bc.Trace.Endpoint == "" {
		return nil, func() {}, nil
	}

	traceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(bc.Trace.Endpoint),
	}
	if bc.Trace.Insecure {
		traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(context.Background(), traceOpts...)
	if err != nil {
		return nil, nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(newResource()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Errorf("failed to shutdown tracer provider: %v", err)
		}
	}

	return tp, cleanup, nil
}
