package guards

import (
	"context"
	"strings"
	"testing"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

// opTransport embeds a nil Transporter and overrides exactly Operation, so
// middleware tests spell out only the behavior under test.
type opTransport struct {
	transport.Transporter
	op string
}

func (t opTransport) Operation() string { return t.op }

type fakeAuthorizer struct {
	decision *kauthz.Decision
}

func (f fakeAuthorizer) Decide(context.Context, *security.Subject, string) (*kauthz.Decision, error) {
	return f.decision, nil
}

// predicateViaMiddleware runs the real authz middleware so the Decision
// reaches ctx the way production injects it, then compiles the mixin
// predicate inside the handler.
func predicateViaMiddleware(t *testing.T, m local_mixins.DatascopeMixin, d *kauthz.Decision) (func(*sql.Selector), bool) {
	t.Helper()
	var (
		pred func(*sql.Selector)
		ok   bool
	)
	ctx := security.WithSubject(
		transport.NewServerContext(context.Background(), opTransport{op: "/cyber.system.v1.UserService/ListUsers"}),
		&security.Subject{UserID: "u1", TenantID: "default"},
	)
	handler := func(ctx context.Context, _ any) (any, error) {
		pred, ok = m.ScopePredicate(ctx), true
		return nil, nil
	}
	if _, err := kauthz.OperationAuthz(fakeAuthorizer{decision: d})(handler)(ctx, nil); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	return pred, ok
}

func renderPred(t *testing.T, pred func(*sql.Selector)) string {
	t.Helper()
	s := sql.Select("*").From(sql.Table("rows"))
	pred(s)
	q, _ := s.Query()
	return q
}

func TestScopePredicatePlumbing(t *testing.T) {
	m := local_mixins.DatascopeMixin{UserField: "id", DeptField: "dept_id"}

	if pred := m.ScopePredicate(context.Background()); pred != nil {
		t.Fatalf("no decision in ctx pred = %v, want false (trusted paths skip filtering)", pred != nil)
	}
	if pred := m.ScopePredicate(local_mixins.Unscoped(context.Background())); pred != nil {
		t.Fatalf("lifted read pred = %v, want false", pred != nil)
	}

	pred, ok := predicateViaMiddleware(t, m, &kauthz.Decision{Allowed: true})
	if !ok || pred != nil {
		t.Fatalf("allowed with empty scopes: pred set = %v, handler ran = %v; want no predicate — narrows nothing", pred != nil, ok)
	}

	pred, ok = predicateViaMiddleware(t, m, &kauthz.Decision{Allowed: true, Scopes: []kauthz.ScopeDescriptor{{Kind: kauthz.ScopeKindSelf}}})
	if !ok || pred == nil {
		t.Fatalf("self scope: pred set = %v, handler ran = %v; want a predicate", pred != nil, ok)
	}
	if q := renderPred(t, pred); !strings.Contains(q, "`id` = ?") {
		t.Fatalf("self predicate sql = %q, want owner equality", q)
	}
}

func TestScopePredicateFailsClosed(t *testing.T) {
	// A self-only narrowing on an entity without UserField has nothing
	// honorable: the empty set, never unfiltered.
	deptMixin := local_mixins.DatascopeMixin{DeptField: "id"}
	pred, ok := predicateViaMiddleware(t, deptMixin, &kauthz.Decision{
		Allowed: true,
		Scopes:  []kauthz.ScopeDescriptor{{Kind: kauthz.ScopeKindSelf}},
	})
	if !ok || pred == nil {
		t.Fatalf("inexpressible narrowing: pred set = %v, handler ran = %v; want a fail-closed predicate", pred != nil, ok)
	}
	if q := renderPred(t, pred); !strings.Contains(strings.ToLower(q), "false") {
		t.Fatalf("fail-closed sql = %q, want the false constant", q)
	}
}
