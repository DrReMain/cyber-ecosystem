package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// buildLogger composes the configured sinks (console/file/otlp) into one fanout
// logger gated by a single level.
//
// Service identity (service.name/version/id) and trace correlation
// (trace_id/span_id) are injected ONLY into the local sinks (console/file).
// They are plain slog handlers with no resource model and no OTel auto trace
// extraction, so they need the injection to stay readable. The OTLP sink does
// NOT receive these injected attrs — that would duplicate them, because:
//   - service.* already arrives via the shared OTel Resource (the resources map);
//   - trace correlation is filled by otelslog itself from the record's context
//     into the LogRecord's top-level trace_id/span_id fields.
//
// The fanout handler is therefore a pure level gate + multiplexer: it injects
// nothing. This avoids the duplicate-fields problem in SigNoz, where the same
// key (service.*, trace_id, span_id) previously landed in both the record's
// attributes and its resources / top-level trace fields.
func buildLogger(cfg Config, name, version, instanceID string, res *resource.Resource, shutdowns *[]func(context.Context) error) *slog.Logger {
	level := parseLevel(cfg.Log.Level)
	svcAttrs := []slog.Attr{
		slog.String("service.name", name),
		slog.String("service.version", version),
		slog.String("service.id", instanceID),
	}
	var handlers []slog.Handler

	// console/file carry no level option — the fanout gates all sinks uniformly
	// by `level`. Each is wrapped in a localSink so records gain trace correlation
	// from their context; service identity is baked in via WithAttrs.
	if cfg.Log.Console {
		base := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})
		handlers = append(handlers, &localSink{inner: base.WithAttrs(svcAttrs)})
	}
	if cfg.Log.File != nil && cfg.Log.File.Path != "" {
		w := &lumberjack.Logger{
			Filename:   cfg.Log.File.Path,
			MaxSize:    cfg.Log.File.MaxSizeMB,
			MaxBackups: cfg.Log.File.MaxBackups,
			MaxAge:     cfg.Log.File.MaxAgeDays,
			Compress:   cfg.Log.File.Compress,
		}
		base := slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true})
		handlers = append(handlers, &localSink{inner: base.WithAttrs(svcAttrs)})
	}
	if cfg.Log.OTLP && cfg.Endpoint != "" {
		opts := []otlploghttp.Option{otlploghttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		if exp, err := otlploghttp.New(context.Background(), opts...); err == nil {
			lp := sdklog.NewLoggerProvider(
				sdklog.WithResource(res),
				sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
			)
			otellogglobal.SetLoggerProvider(lp)
			// otelslog pulls service identity from the provider's resource and fills
			// trace correlation from the record's context — no manual injection.
			handlers = append(handlers, otelslog.NewHandler(name, otelslog.WithLoggerProvider(lp)))
			*shutdowns = append(*shutdowns, lp.Shutdown)
		}
	}

	return slog.New(&fanoutHandler{level: level, handlers: handlers})
}

// localSink wraps a plain slog handler (console/file) to inject trace_id/span_id
// from the record's context. Local sinks have no OTel auto-extraction, so this is
// how they gain trace correlation. Service identity is baked in at construction
// via WithAttrs (see buildLogger). The OTLP sink skips this wrapper on purpose.
type localSink struct{ inner slog.Handler }

func (s *localSink) Enabled(ctx context.Context, l slog.Level) bool { return s.inner.Enabled(ctx, l) }

func (s *localSink) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return s.inner.Handle(ctx, r)
}

func (s *localSink) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &localSink{inner: s.inner.WithAttrs(attrs)}
}

func (s *localSink) WithGroup(name string) slog.Handler {
	return &localSink{inner: s.inner.WithGroup(name)}
}

// fanoutHandler forwards each record to multiple slog handlers and is the single
// level gate (so every sink — including OTLP — respects the configured level). It
// injects nothing: per-sink concerns (service identity, trace correlation) live
// on each sink — localSink for console/file, otelslog + resource for OTLP.
type fanoutHandler struct {
	level    slog.Level
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hs := range h.handlers {
		if !hs.Enabled(ctx, r.Level) {
			continue
		}
		// Clone per handler: localSink mutates the record (adds trace attrs).
		if err := hs.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, x := range h.handlers {
		hs[i] = x.WithAttrs(attrs)
	}
	return &fanoutHandler{level: h.level, handlers: hs}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, x := range h.handlers {
		hs[i] = x.WithGroup(name)
	}
	return &fanoutHandler{level: h.level, handlers: hs}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
