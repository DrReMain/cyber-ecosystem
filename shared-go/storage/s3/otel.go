package s3

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func startSpan(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.GetTracerProvider().Tracer("storage")
	allAttrs := []attribute.KeyValue{
		attribute.String("storage.backend", "s3"),
		attribute.String("storage.operation", op),
	}
	allAttrs = append(allAttrs, attrs...)
	ctx, span := tracer.Start(ctx, "storage "+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(allAttrs...),
	)
	return ctx, span
}

func recordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
