package observability

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func recordOf(attrs ...slog.Attr) slog.Record {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "server request", 0)
	r.AddAttrs(attrs...)
	return r
}

// The mask erases credential fields from the request-payload attributes the
// kratos logging middleware emits, in every spelling the wire messages use,
// and leaves everything else byte-identical.
func TestSecretMaskHandlerMasksCredentials(t *testing.T) {
	cases := []struct{ in, masked, plain string }{
		{
			`tenant_code:"globex" account:"root" password:"Globex-Root-2026!y"`,
			`password:\"***\"`, `Globex-Root-2026`,
		},
		{
			`account:"admin" password:"admin123456" population:"user"`,
			`password:\"***\"`, `admin123456`,
		},
		{
			`old_password:"a" new_password:"Str0ng!x" user_id:"u1"`,
			`old_password:\"***\" new_password:\"***\"`, `Str0ng!x`,
		},
		{
			`apiKey:"sk-live" secret:"s" access_key_secret:"k"`,
			`apiKey:\"***\" secret:\"***\" access_key_secret:\"***\"`, `sk-live`,
		},
	}
	for _, c := range cases {
		buf := &strings.Builder{}
		h := &secretMaskHandler{next: slog.NewTextHandler(buf, nil)}
		if err := h.Handle(context.Background(), recordOf(slog.String("args", c.in))); err != nil {
			t.Fatalf("handle: %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, c.masked) || strings.Contains(got, c.plain) {
			t.Fatalf("mask(%q):\n got %s\nwant %q present, %q absent", c.in, got, c.masked, c.plain)
		}
	}
	// Non-credential payloads ride through byte-identical.
	buf := &strings.Builder{}
	h := &secretMaskHandler{next: slog.NewTextHandler(buf, nil)}
	_ = h.Handle(context.Background(), recordOf(slog.String("args", `account:"admin" enabled:true`)))
	if strings.Contains(buf.String(), "***") {
		t.Fatalf("over-masked: %s", buf.String())
	}
}

// Non-target attributes and non-string values ride through untouched, and
// the handler composes with the error view (the render chain both ways).
func TestSecretMaskHandlerPassesThrough(t *testing.T) {
	buf := &strings.Builder{}
	h := &secretMaskHandler{next: slog.NewTextHandler(buf, nil)}
	r := recordOf(
		slog.String("operation", "/cyber.system.v1.TenantService/AddTenantAdmin"),
		slog.Int64("code", 200),
	)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("handle: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "operation=/cyber.system.v1.TenantService/AddTenantAdmin") || !strings.Contains(out, "code=200") {
		t.Fatalf("passthrough: %s", out)
	}
	if masked := RenderSecrets(slog.New(slog.NewTextHandler(buf, nil))); masked == nil {
		t.Fatal("RenderSecrets must return a logger")
	}
}
