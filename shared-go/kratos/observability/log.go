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
// logger. The fanout gates all sinks by a single level and injects
// trace_id/span_id once — but only when the record's context actually carries a
// span, so non-request (startup/background) logs stay clean. Appends any OTLP
// log provider shutdown to *shutdowns.
func buildLogger(cfg Config, name, version, instanceID string, res *resource.Resource, shutdowns *[]func(context.Context) error) *slog.Logger {
	level := parseLevel(cfg.Log.Level)
	var handlers []slog.Handler

	// console/file handlers carry no level option — the fanout gates all sinks
	// uniformly by `level`.
	if cfg.Log.Console {
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	}
	if cfg.Log.File != nil && cfg.Log.File.Path != "" {
		w := &lumberjack.Logger{
			Filename:   cfg.Log.File.Path,
			MaxSize:    cfg.Log.File.MaxSizeMB,
			MaxBackups: cfg.Log.File.MaxBackups,
			MaxAge:     cfg.Log.File.MaxAgeDays,
			Compress:   cfg.Log.File.Compress,
		}
		handlers = append(handlers, slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true}))
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
			handlers = append(handlers, otelslog.NewHandler(name, otelslog.WithLoggerProvider(lp)))
			*shutdowns = append(*shutdowns, lp.Shutdown)
		}
	}

	return slog.New(&fanoutHandler{level: level, handlers: handlers}).
		With("service.name", name, "service.version", version, "service.id", instanceID)
}

// fanoutHandler forwards each record to multiple slog handlers. It is the single
// level gate (so every sink — including OTLP — respects the configured level)
// and the single trace-correlation injection point: only request-scoped records
// (those whose context carries a span) get trace_id/span_id.
type fanoutHandler struct {
	level    slog.Level
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		r = r.Clone()
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	for _, hs := range h.handlers {
		if !hs.Enabled(ctx, r.Level) {
			continue
		}
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
