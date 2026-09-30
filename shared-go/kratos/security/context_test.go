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
