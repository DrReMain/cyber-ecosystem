// Package observability provides backend instrumentation helpers that attach
// OTel tracing/metrics to the auto-instrumented backends (redis, s3, database/sql)
// and to MQ (manual decorator). Every helper reads the GLOBAL OTel providers set
// by Init (otel.SetTracerProvider / otel.SetMeterProvider) — none of them take a
// provider argument, so the platform layer needs no per-backend wiring beyond
// calling these once after Init and passing the returned option/opener through.
//
// Injection is deliberately split: the backend packages themselves (storage/s3,
// orm/ent/client) stay free of observability imports. The platform layer passes
// the helper's result in (an s3.Options func / a SQLOpener closure), keeping the
// dependency direction one-way.
package observability

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"

	"cyber-ecosystem/shared-go/capability/mq"
)

// dbAttr is the static db.system attribute reused across SQL helpers. Pinned to
// postgresql since every backend in this repo targets Postgres; callers wanting
// another system should add a dedicated helper rather than parameterize here.
var dbAttr = semconv.DBSystemPostgreSQL

// InstrumentRedis attaches OTel tracing + metrics to a redis client. Both
// redisotel helpers default to the global providers (no option needed). Returns
// a combined error so callers can fail fast on a single setup mistake.
func InstrumentRedis(c *redis.Client) error {
	if err := redisotel.InstrumentTracing(c); err != nil {
		return fmt.Errorf("redis trace instrumentation: %w", err)
	}
	if err := redisotel.InstrumentMetrics(c); err != nil {
		return fmt.Errorf("redis metrics instrumentation: %w", err)
	}
	return nil
}

// S3Options returns an s3.Options mutator that appends the aws-sdk-go-v2 OTel
// middlewares (trace + metrics) to the client's APIOptions. Pass it to
// storage/s3.NewClient via its variadic opts; storage/s3 stays unaware of
// observability. otelaws.AppendMiddlewares reads the global providers.
func S3Options() func(*s3.Options) {
	return func(o *s3.Options) {
		otelaws.AppendMiddlewares(&o.APIOptions)
	}
}

// OpenSQL wraps a database/sql driver with OTel tracing + connection-pool
// metrics via otelsql. Flow: register a traced driver name for baseDriver →
// sql.Open that wrapped name against dsn → register DBStats metrics on the
// resulting *sql.DB. The returned *sql.DB is what ent/orm should hold.
//
// The metric.Registration from RegisterDBStatsMetrics is intentionally not
// returned: it is tied to the *sql.DB lifetime and the caller closes the db
// directly; an explicit Unregister hook would just duplicate that lifecycle.
// otelsql defaults to the global tracer/meter providers.
func OpenSQL(baseDriver, dsn string) (*sql.DB, error) {
	wrappedName, err := otelsql.Register(baseDriver, otelsql.WithAttributes(dbAttr))
	if err != nil {
		return nil, fmt.Errorf("otelsql register %q: %w", baseDriver, err)
	}
	db, err := sql.Open(wrappedName, dsn)
	if err != nil {
		return nil, fmt.Errorf("otelsql open: %w", err)
	}
	if _, err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(dbAttr)); err != nil {
		// A failed metrics registration must not leak the opened handle.
		_ = db.Close()
		return nil, fmt.Errorf("otelsql db stats metrics: %w", err)
	}
	return db, nil
}

// InstrumentMQ wraps an MQ's Publisher and Consumer with manual spans and W3C
// tracecontext propagation through msg.Headers (a propagation.MapCarrier).
// Publishing injects the current context's trace into Headers; consuming
// extracts it, starting a consumer span under the producer span. Returns a new
// *mq.MQ sharing the same backends — the caller swaps the original for the
// wrapped one. Uses the global tracer + global propagator (W3C by default).
func InstrumentMQ(m *mq.MQ) *mq.MQ {
	tracer := otel.Tracer("cyber-ecosystem/mq")
	return &mq.MQ{
		Publisher: &tracedPublisher{inner: m.Publisher, tracer: tracer},
		Consumer:  &tracedConsumer{inner: m.Consumer, tracer: tracer},
	}
}

type tracedPublisher struct {
	inner  mq.Publisher
	tracer trace.Tracer
}

func (p *tracedPublisher) Publish(ctx context.Context, topic string, msg *mq.Message) (string, error) {
	ctx, span := p.tracer.Start(ctx, "mq.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("messaging.destination.name", topic)),
	)
	defer span.End()

	if msg.Headers == nil {
		msg.Headers = map[string]string{}
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Headers))

	id, err := p.inner.Publish(ctx, topic, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	return id, nil
}

type tracedConsumer struct {
	inner  mq.Consumer
	tracer trace.Tracer
}

func (c *tracedConsumer) Subscribe(ctx context.Context, topic, group string, handler func(context.Context, mq.Message) error) (mq.Subscription, error) {
	return c.inner.Subscribe(ctx, topic, group, func(ctx context.Context, msg mq.Message) error {
		// Extract the producer's trace context from Headers to continue the
		// trace across the queue hop; fall back to a root span when absent.
		parentCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msg.Headers))
		ctx2, span := c.tracer.Start(parentCtx, "mq.process",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(attribute.String("messaging.destination.name", topic)),
		)
		defer span.End()

		if err := handler(ctx2, msg); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return nil
	})
}
