package introspect

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// fakeAuthorizer answers the decide seam with a fixed shape and records
// the ask; the remaining Authorizer surface is not this module's to test.
type fakeAuthorizer struct {
	kauthz.Authorizer
	decision  *kauthz.Decision
	err       error
	subject   *security.Subject
	operation string
	calls     int
}

func (f *fakeAuthorizer) Decide(_ context.Context, s *security.Subject, op string) (*kauthz.Decision, error) {
	f.calls++
	f.subject, f.operation = s, op
	return f.decision, f.err
}

func newIntrospectService(a *fakeAuthorizer) *IntrospectService {
	return NewIntrospectService(slog.New(slog.NewTextHandler(io.Discard, nil)), a)
}

// VerifySession mirrors the authenticated subject the middleware already
// resolved — six fields, nothing computed.
func TestVerifySessionMirrorsSubject(t *testing.T) {
	svc := newIntrospectService(&fakeAuthorizer{})
	subject := &security.Subject{
		UserID: "u1", TenantID: "t1", SessionID: "s1", Population: "member", Plane: "member",
		Extra: map[string]any{security.ExtraAccount: "acct"},
	}

	out, err := svc.VerifySession(security.WithSubject(context.Background(), subject), &systempb.VerifySessionRequest{})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if out.GetUserId().GetValue() != "u1" || out.GetTenantId().GetValue() != "t1" ||
		out.GetSessionId().GetValue() != "s1" || out.GetAccount().GetValue() != "acct" ||
		out.GetPopulation().GetValue() != "member" || out.GetPlane().GetValue() != "member" {
		t.Fatalf("mirror: %+v", out)
	}
}

// Both faces gate on the middleware's subject before anything else.
func TestRequiresSubject(t *testing.T) {
	svc := newIntrospectService(&fakeAuthorizer{})
	bare := context.Background()

	if _, err := svc.VerifySession(bare, &systempb.VerifySessionRequest{}); !errors.Is(err, kauthz.ErrMissingSubject) {
		t.Fatalf("verify without subject: want MISSING_SUBJECT, got %v", err)
	}
	a := &fakeAuthorizer{}
	if _, err := newIntrospectService(a).CheckOperation(bare, &systempb.CheckOperationRequest{Operation: "/op"}); !errors.Is(err, kauthz.ErrMissingSubject) {
		t.Fatalf("check without subject: want MISSING_SUBJECT, got %v", err)
	}
	if a.calls != 0 {
		t.Fatal("check without subject: the engine must not be consulted")
	}
}

// CheckOperation forwards the operation with the subject and answers the
// decision's own verdict.
func TestCheckOperationAnswersDecision(t *testing.T) {
	a := &fakeAuthorizer{decision: &kauthz.Decision{Allowed: true}}
	svc := newIntrospectService(a)
	subject := &security.Subject{UserID: "u1", TenantID: "t1"}

	out, err := svc.CheckOperation(security.WithSubject(context.Background(), subject), &systempb.CheckOperationRequest{Operation: "/cyber.Svc/Op"})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !out.GetAllowed() {
		t.Fatalf("allowed: %+v", out)
	}
	if a.subject != subject || a.operation != "/cyber.Svc/Op" {
		t.Fatalf("ask: %+v %q", a.subject, a.operation)
	}
}

// A nil decision is the engine's deny-with-no-reason shape — the face
// answers the uniform forbidden, never a zero-value response.
func TestCheckOperationNilDecisionDenies(t *testing.T) {
	svc := newIntrospectService(&fakeAuthorizer{})
	subject := &security.Subject{UserID: "u1", TenantID: "t1"}

	if _, err := svc.CheckOperation(security.WithSubject(context.Background(), subject), &systempb.CheckOperationRequest{Operation: "/op"}); !errors.Is(err, kauthz.ErrPermissionDenied) {
		t.Fatalf("nil decision: want PERMISSION_DENIED, got %v", err)
	}
}

// An engine failure propagates — the introspection face adds no policy
// of its own on top of the engine.
func TestCheckOperationPropagatesEngineError(t *testing.T) {
	boom := errors.New("boom")
	svc := newIntrospectService(&fakeAuthorizer{err: boom})
	subject := &security.Subject{UserID: "u1", TenantID: "t1"}

	if _, err := svc.CheckOperation(security.WithSubject(context.Background(), subject), &systempb.CheckOperationRequest{Operation: "/op"}); !errors.Is(err, boom) {
		t.Fatalf("engine error: got %v", err)
	}
}
