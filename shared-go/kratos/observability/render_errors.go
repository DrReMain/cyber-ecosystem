package observability

import (
	"context"
	stderrors "errors"
	"log/slog"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// RenderErrors returns a logger whose handler flattens kratos error chains in
// log records: the "error" attribute becomes the factual root (identity
// already lives in the dedicated code/reason fields), the duplicate "stack"
// attribute is dropped, a non-empty client-facing message is surfaced as its
// own attribute, and expected (4xx) failures are demoted to WARN. Hand this
// logger to the stock kratos logging middleware — it renders *errors.Error
// via Error(), which inlines the whole chain recursively and writes it twice
// (error + stack). Rewriting at the handler keeps the middleware untouched:
// if its attribute shape ever changes upstream, the worst case is falling
// back to the stock rendering.
func RenderErrors(logger *slog.Logger) *slog.Logger {
	return slog.New(&errorViewHandler{next: logger.Handler()})
}

// errorViewHandler forwards records to next, rewriting the error projection of
// records whose "error" attribute carries a kratos error. Non-kratos records
// pass through unchanged.
type errorViewHandler struct{ next slog.Handler }

func (h *errorViewHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *errorViewHandler) Handle(ctx context.Context, r slog.Record) error {
	view := errorViewOf(&r)
	if view == nil {
		return h.next.Handle(ctx, r)
	}

	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	// Client-side and rule violations (4xx) are expected events; only server
	// faults deserve the ERROR level and its alerting weight.
	if r.Level >= slog.LevelError && view.code > 0 && view.code < 500 {
		nr.Level = slog.LevelWarn
	}
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "stack":
			// Dropped: for a kratos error the middleware derives it from the
			// same %+v rendering as the error attribute — pure duplication.
		case "error":
			if view.message != "" {
				nr.AddAttrs(slog.String("message", view.message))
			}
			if path := view.internalPath(); path != "" {
				nr.AddAttrs(slog.String("internal", path))
			}
			if view.root != "" {
				nr.AddAttrs(slog.String("error", view.root))
			}
		default:
			nr.AddAttrs(a)
		}
		return true
	})
	// Re-check the level gate after demotion: the record already passed
	// Enabled at its original level, and a WARN demotion under a configured
	// error-only level should not reach the sinks.
	if !h.next.Enabled(ctx, nr.Level) {
		return nil
	}
	return h.next.Handle(ctx, nr)
}

// WithAttrs forwards verbatim: the rewrite targets the per-record attributes
// the stock kratos logging middleware emits (error/stack arrive via LogAttrs,
// never via With), so pre-bound attributes carry no chain to re-project.
func (h *errorViewHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &errorViewHandler{next: h.next.WithAttrs(attrs)}
}

func (h *errorViewHandler) WithGroup(name string) slog.Handler {
	return &errorViewHandler{next: h.next.WithGroup(name)}
}

// errorViewOf returns the flattened view of the kratos error carried by the
// record's "error" attribute, or nil when the record carries none.
func errorViewOf(r *slog.Record) *errorView {
	var view *errorView
	r.Attrs(func(a slog.Attr) bool {
		if view != nil || a.Key != "error" {
			return true
		}
		err, ok := a.Value.Any().(error)
		if !ok {
			return true
		}
		var ke *kratoserrors.Error
		if stderrors.As(err, &ke) {
			v := viewError(err)
			view = &v
		}
		return true
	})
	return view
}
