package authz

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// fakeAuthzRP answers the engine slice with a fixed grant set — or the
// injected engine error; implicitCalls pins the "patron never consults the
// engine" law.
type fakeAuthzRP struct {
	AuthzRP
	grants        []GrantView
	err           error
	implicitCalls int
	patron        bool
	patronErr     error
}

func (f *fakeAuthzRP) ImplicitGrants(context.Context, string, string) ([]GrantView, error) {
	f.implicitCalls++
	return f.grants, f.err
}

func (f *fakeAuthzRP) PatronSubject(subject *security.Subject) (bool, error) {
	// Mirrors the port: an empty population is an engine subject.
	if subject.Population == "" {
		return false, nil
	}
	if f.patronErr != nil {
		return false, f.patronErr
	}
	return f.patron, nil
}

// hookSink discards lifecycle hooks — the watch loop is not under test here.
type hookSink struct{}

func (hookSink) OnStart(func(context.Context) error) {}
func (hookSink) OnStop(func(context.Context) error)  {}

func newAuthzUC(rp *fakeAuthzRP) *AuthzUC {
	return NewAuthzUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, rp, hookSink{},
	)
}

// Decide --------------------------------------------------------------------------------------------------------------

// Patron subjects are structurally grantless — identity is their whole
// permission; the engine is never consulted.
func TestDecidePatronNeverConsultsEngine(t *testing.T) {
	rp := &fakeAuthzRP{grants: []GrantView{{Pattern: "/*"}}}
	uc := newAuthzUC(&fakeAuthzRP{patron: true})

	d, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1", Population: "guest"}, "system:user:list")
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Allowed || d.Reason != kauthz.DenyNoGrant {
		t.Fatalf("patron: decision = %+v, want deny/no-grant", d)
	}
	if rp.implicitCalls != 0 {
		t.Fatal("patron: the engine must not be consulted")
	}

	// A snapshot failure surfaces as itself — never a deny that would read
	// as policy.
	uc2 := newAuthzUC(&fakeAuthzRP{patronErr: errorspb.ErrorGeneralErrorInternal("")})
	if _, err := uc2.Decide(context.Background(), &security.Subject{UserID: "u1", Population: "guest"}, "op"); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("snapshot failure: want INTERNAL, got %v", err)
	}
}

// Engine failure is infra, not a domain rule — INTERNAL, never a deny that
// would read as policy.
func TestDecideEngineFailureIsInternal(t *testing.T) {
	uc := newAuthzUC(&fakeAuthzRP{err: errorspb.ErrorInfraErrorNetworkConnection("")})

	if _, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1"}, "/cyber.system.v1.UserService/ListUsers"); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("engine failure: want INTERNAL, got %v", err)
	}
}

// A grant that does not match the operation contributes nothing — no grant,
// no allow.
func TestDecideNoMatchingGrantDenies(t *testing.T) {
	uc := newAuthzUC(&fakeAuthzRP{grants: []GrantView{{Pattern: "/cyber.other.Svc/Method"}}})

	d, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1"}, "/cyber.system.v1.UserService/ListUsers")
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Allowed || d.Reason != kauthz.DenyNoGrant {
		t.Fatalf("no match: decision = %+v, want denied/no_grant", d)
	}
}

// An unregistered policy kind fails closed: the grant is eliminated, and the
// deny carries the ABAC reason, distinct from a plain no-grant.
func TestDecideUnregisteredPolicyFailsClosed(t *testing.T) {
	uc := newAuthzUC(&fakeAuthzRP{grants: []GrantView{{
		Pattern:  "/cyber.system.v1.UserService/*",
		Policies: []PolicyView{{Kind: "no-such-kind", Name: "n"}},
	}}})

	d, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1"}, "/cyber.system.v1.UserService/ListUsers")
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Allowed || d.Reason != kauthz.DenyAbacConstraint {
		t.Fatalf("fail-closed: decision = %+v, want denied/abac_constraint", d)
	}
}

// A matching grant without linked policies passes as-is — zero attribute
// resolution, scopes derived from the grant.
func TestDecideGrantWithoutPolicyPasses(t *testing.T) {
	uc := newAuthzUC(&fakeAuthzRP{grants: []GrantView{{Pattern: "/cyber.system.v1.UserService/*"}}})

	d, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1"}, "/cyber.system.v1.UserService/ListUsers")
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !d.Allowed || len(d.Scopes) != 1 {
		t.Fatalf("plain grant: decision = %+v, want allowed with one scope", d)
	}
}

// matchOperation ------------------------------------------------------------------------------------------------------

// Match semantics mirror the client verbatim: exact for non-wildcards, prefix
// for a clean "/*" tail; a non-tail "*" is not a wildcard.
func TestMatchOperationSemantics(t *testing.T) {
	cases := []struct {
		op, pattern string
		want        bool
	}{
		{"/cyber.s.M/a", "/cyber.s.M/a", true},
		{"/cyber.s.M/a", "/cyber.s.M/b", false},
		{"/cyber.s.M/a", "/cyber.s.M/*", true},
		{"/cyber.t.N/a", "/cyber.s.M/*", false},
		{"/cyber.s.M/ab", "/cyber.s.M/a*", false},
		{"/cyber.s.M/*", "/*", true},
	}
	for _, c := range cases {
		if got := matchOperation(c.op, c.pattern); got != c.want {
			t.Fatalf("match(%q, %q) = %v, want %v", c.op, c.pattern, got, c.want)
		}
	}
}
