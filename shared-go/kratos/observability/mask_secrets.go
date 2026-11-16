package observability

import (
	"context"
	"log/slog"
	"regexp"
)

// secretFieldRe matches credential-shaped fields in the proto-text rendering
// the stock kratos logging middleware writes into the "args" attribute (and
// error chains render the same shape). Prototext emits json names, so both
// spellings of every field must match: password/new_password/oldPassword,
// apiKey/api_key, accessKeySecret/access_key_secret, secret. Login, password
// changes, and tenant admin provisioning all carry plaintext credentials on
// the wire (TLS); the request log must not become a second, quieter copy.
var secretFieldRe = regexp.MustCompile(`(?i)\b((?:old_?|new_?)?password|access_?key_?secret|secret|api_?key):"[^"]*"`)

// RenderSecrets returns a logger whose handler masks credential fields in
// the string attributes the kratos logging middleware emits (args, error,
// message). Non-string attributes and other keys pass through untouched.
func RenderSecrets(logger *slog.Logger) *slog.Logger {
	return slog.New(&secretMaskHandler{next: logger.Handler()})
}

// secretMaskHandler rewrites the request-payload-bearing attributes of each
// record; it composes with errorViewHandler (either order works — both are
// per-record, stateless rewrites).
type secretMaskHandler struct{ next slog.Handler }

func (h *secretMaskHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *secretMaskHandler) Handle(ctx context.Context, r slog.Record) error {
	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "args", "error", "message":
			if s := a.Value.String(); s != "" {
				nr.AddAttrs(slog.String(a.Key, secretFieldRe.ReplaceAllString(s, `$1:"***"`)))
				return true
			}
		}
		nr.AddAttrs(a)
		return true
	})
	return h.next.Handle(ctx, nr)
}

func (h *secretMaskHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &secretMaskHandler{next: h.next.WithAttrs(attrs)}
}

func (h *secretMaskHandler) WithGroup(name string) slog.Handler {
	return &secretMaskHandler{next: h.next.WithGroup(name)}
}
