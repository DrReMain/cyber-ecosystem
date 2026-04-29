package zap

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sourcesdk "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

type OtlpLogConfig struct {
	Enabled        bool
	Endpoint       string
	Insecure       bool
	ServiceName    string
	ServiceID      string
	ServiceVersion string
}

type OtlpLogWriter struct {
	provider *sdklog.LoggerProvider
	logger   otellog.Logger
}

func NewOtlpLogWriter(cfg *OtlpLogConfig) (*OtlpLogWriter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "unknown"
	}

	resOpts := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
	}
	if cfg.ServiceID != "" {
		resOpts = append(resOpts, semconv.ServiceInstanceIDKey.String(cfg.ServiceID))
	}
	if cfg.ServiceVersion != "" {
		resOpts = append(resOpts, semconv.ServiceVersionKey.String(cfg.ServiceVersion))
	}
	res := sourcesdk.NewSchemaless(resOpts...)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	return &OtlpLogWriter{
		provider: provider,
		logger:   provider.Logger(serviceName),
	}, nil
}

func (w *OtlpLogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	var raw map[string]any
	if json.Unmarshal(p, &raw) != nil {
		return len(p), nil
	}

	var record otellog.Record

	if ts, ok := raw["ts"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			record.SetTimestamp(t)
		}
	}

	if level, ok := raw["level"].(string); ok {
		record.SetSeverity(parseOtelSeverity(level))
		record.SetSeverityText(level)
	}

	if msg, ok := raw["msg"].(string); ok {
		record.SetBody(otellog.StringValue(msg))
	}

	emitCtx := context.Background()

	if traceIDStr, _ := raw["trace.id"].(string); traceIDStr != "" {
		if tid, err := trace.TraceIDFromHex(traceIDStr); err == nil {
			if spanIDStr, _ := raw["span.id"].(string); spanIDStr != "" {
				if sid, err := trace.SpanIDFromHex(spanIDStr); err == nil {
					sc := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    tid,
						SpanID:     sid,
						TraceFlags: trace.FlagsSampled,
					})
					emitCtx = trace.ContextWithSpanContext(emitCtx, sc)
				}
			}
		}
	}

	attrs := make([]otellog.KeyValue, 0, len(raw))
	for k, v := range raw {
		switch k {
		case "ts", "level", "msg", "trace.id", "span.id", "service.name", "service.id", "service.version":
			continue
		}
		attrs = append(attrs, otelLogKV(k, v))
	}
	if len(attrs) > 0 {
		record.AddAttributes(attrs...)
	}

	w.logger.Emit(emitCtx, record)
	return len(p), nil
}

func (w *OtlpLogWriter) Sync() error {
	return w.provider.ForceFlush(context.Background())
}

func (w *OtlpLogWriter) Close() error {
	return w.provider.Shutdown(context.Background())
}

func parseOtelSeverity(level string) otellog.Severity {
	switch level {
	case "debug":
		return otellog.SeverityDebug
	case "info":
		return otellog.SeverityInfo
	case "warn":
		return otellog.SeverityWarn
	case "error":
		return otellog.SeverityError
	case "dpanic":
		return otellog.SeverityError
	case "panic":
		return otellog.SeverityFatal4
	case "fatal":
		return otellog.SeverityFatal
	default:
		return otellog.SeverityInfo
	}
}

func otelLogKV(k string, v any) otellog.KeyValue {
	switch val := v.(type) {
	case string:
		return otellog.String(k, val)
	case float64:
		return otellog.Float64(k, val)
	case int:
		return otellog.Int(k, val)
	case int64:
		return otellog.Int64(k, val)
	case uint:
		return otellog.Int64(k, int64(val))
	case uint64:
		return otellog.Int64(k, int64(val))
	case bool:
		return otellog.Bool(k, val)
	case []byte:
		return otellog.String(k, string(val))
	default:
		return otellog.String(k, fmt.Sprintf("%v", val))
	}
}
