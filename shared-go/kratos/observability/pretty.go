package observability

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PrettyProto returns a logger whose handler renders any protobuf-message log
// attribute as readable JSON (via protojson) instead of the raw Go struct dump
// slog emits by default (e.g. kratos recovery's "request" attr: a noisy
// `{state:{NoUnkeyedLiterals:...}}`). Non-proto attributes pass through
// unchanged. Use it to wrap a logger handed to code that logs raw request
// payloads, e.g. recovery.Recovery(recovery.WithLogger(observability.PrettyProto(logger))).
func PrettyProto(logger *slog.Logger) *slog.Logger {
	return slog.New(&protoJSONHandler{next: logger.Handler()})
}

// protoJSONHandler forwards records to next, rewriting any attribute whose value
// is a proto.Message into its protojson form. It only rewrites proto values;
// strings (stack, trace_id, ...) are untouched.
type protoJSONHandler struct{ next slog.Handler }

func (h *protoJSONHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *protoJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(prettifyProtoAttr(a))
		return true
	})
	return h.next.Handle(ctx, nr)
}

func (h *protoJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		out[i] = prettifyProtoAttr(a)
	}
	return &protoJSONHandler{next: h.next.WithAttrs(out)}
}

func (h *protoJSONHandler) WithGroup(name string) slog.Handler {
	return &protoJSONHandler{next: h.next.WithGroup(name)}
}

func prettifyProtoAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() != slog.KindAny {
		return a
	}
	v := a.Value.Any()
	if v == nil {
		return a
	}
	msg, ok := v.(proto.Message)
	if !ok {
		return a
	}
	b, err := protojson.Marshal(msg)
	if err != nil {
		return a
	}
	return slog.String(a.Key, string(b))
}
