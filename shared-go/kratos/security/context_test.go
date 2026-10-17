package security_test

import (
	"context"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
)

func TestSubjectCtx(t *testing.T) {
	if _, ok := security.SubjectFromCtx(context.Background()); ok {
		t.Fatal("empty ctx must not carry a subject")
	}
	subject := &security.Subject{UserID: "u1", TenantID: "default", SessionID: "tok", Extra: map[string]any{"email": "a@b.c"}}
	got, ok := security.SubjectFromCtx(security.WithSubject(context.Background(), subject))
	if !ok || got != subject {
		t.Fatalf("subject = %v, %v; want the stored subject", got, ok)
	}
	if got.Extra["email"] != "a@b.c" {
		t.Fatalf("extra map not preserved: %v", got.Extra)
	}
}

func TestTenantFromCtx(t *testing.T) {
	if got := security.TenantFromCtx(context.Background()); got != security.DefaultTenant {
		t.Fatalf("no subject: tenant = %q, want %q", got, security.DefaultTenant)
	}
	if got := security.TenantFromCtx(security.WithSubject(context.Background(), &security.Subject{UserID: "u1"})); got != security.DefaultTenant {
		t.Fatalf("empty tenant: tenant = %q, want %q", got, security.DefaultTenant)
	}
	if got := security.TenantFromCtx(security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "acme"})); got != "acme" {
		t.Fatalf("tenant = %q, want %q", got, "acme")
	}
}

// The explicit domain switch outranks the subject: row routing (the tenant
// mixin) and bucket routing (per-tenant storage) resolve tenancy through
// this one precedence, so a pass over a domain hits that domain's rows and
// that domain's bucket regardless of whose subject rides the context.
func TestExplicitTenantCtx(t *testing.T) {
	ctx := security.WithTenant(context.Background(), "acme")
	if got, ok := security.ExplicitTenantFromCtx(ctx); !ok || got != "acme" {
		t.Fatalf("explicit tenant = %q, %v; want acme", got, ok)
	}
	if got := security.TenantFromCtx(ctx); got != "acme" {
		t.Fatalf("tenant = %q, want acme", got)
	}
	sub := security.WithSubject(ctx, &security.Subject{UserID: "u1", TenantID: "default"})
	if got := security.TenantFromCtx(sub); got != "acme" {
		t.Fatalf("explicit vs subject: tenant = %q, want acme", got)
	}
	if got, ok := security.ExplicitTenantFromCtx(security.WithTenant(context.Background(), "")); ok {
		t.Fatalf("empty switch: explicit tenant = %q, %v; want unset", got, ok)
	}
}
