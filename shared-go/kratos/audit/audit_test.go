package audit_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	// The named import registers the cyber.* descriptors so actionOf resolves
	// real operations instead of hitting the WRITE floor.
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// sinkStub captures events without any carrier.
type sinkStub struct {
	mu     sync.Mutex
	events []*audit.Event
}

func (s *sinkStub) Emit(_ context.Context, ev *audit.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, ev)
}

func (s *sinkStub) collected() []*audit.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*audit.Event(nil), s.events...)
}

// fakeHeader is a minimal transport.Header over a plain map.
type fakeHeader map[string]string

func (h fakeHeader) Get(key string) string      { return h[key] }
func (h fakeHeader) Set(key, val string)        { h[key] = val }
func (h fakeHeader) Add(key, val string)        { h[key] += "," + val }
func (h fakeHeader) Keys() []string             { return nil }
func (h fakeHeader) Values(key string) []string { return nil }

// fakeTransport covers the generic middleware path (operation + headers).
type fakeTransport struct {
	operation string
	header    fakeHeader
}

func (t *fakeTransport) Kind() transport.Kind            { return transport.KindGRPC }
func (t *fakeTransport) Endpoint() string                { return "" }
func (t *fakeTransport) Operation() string               { return t.operation }
func (t *fakeTransport) RequestHeader() transport.Header { return t.header }
func (t *fakeTransport) ReplyHeader() transport.Header   { return t.header }

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuditCollectionRule(t *testing.T) {
	const (
		writeOp = "/cyber.system.v1.UserService/CreateUser"
		readOp  = "/cyber.system.v1.UserService/ListUsers"
	)
	// The rebound shape: a mapped error carrying the original sentinel as
	// its cause, exactly what the app wiring produces.
	denied := errorspb.ErrorGeneralErrorPermissionDenied("").WithCause(kauthz.ErrPermissionDenied)
	policyDenied := systempb.ErrorSystemAuthzPolicyDenied("").WithCause(kauthz.ErrPolicyDenied)

	tests := []struct {
		name      string
		op        string
		handler   error
		collected bool
		wantDeny  string
	}{
		{"write allow", writeOp, nil, true, ""},
		{"write business failure still collected", writeOp, errorspb.ErrorGeneralErrorValidationFailed(""), true, ""},
		{"read allow not collected", readOp, nil, false, ""},
		{"read deny collected with reason", readOp, denied, true, audit.DenyNoGrant},
		{"abac deny classified separately", readOp, policyDenied, true, audit.DenyAbacConstraint},
		{"non-authz failure on read not collected", readOp, errorspb.ErrorGeneralErrorInternal(""), false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &sinkStub{}
			handler := audit.Audit(sink)(func(_ context.Context, _ any) (any, error) { return nil, tt.handler })
			ctx := transport.NewServerContext(context.Background(), &fakeTransport{operation: tt.op})
			if _, err := handler(ctx, nil); !errors.Is(err, tt.handler) {
				t.Fatalf("handler error mutated: %v", err)
			}
			got := sink.collected()
			if !tt.collected {
				if len(got) != 0 {
					t.Fatalf("collected %d events, want none", len(got))
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("collected %d events, want 1", len(got))
			}
			if got[0].DenyReason != tt.wantDeny {
				t.Errorf("deny reason = %q, want %q", got[0].DenyReason, tt.wantDeny)
			}
			if got[0].Operation != tt.op {
				t.Errorf("operation = %q, want %q", got[0].Operation, tt.op)
			}
		})
	}
}

func TestAuditEventFields(t *testing.T) {
	sink := &sinkStub{}
	handler := audit.Audit(sink)(func(_ context.Context, _ any) (any, error) { return nil, nil })

	tr := &fakeTransport{
		operation: "/cyber.system.v1.AuthService/Login",
		header: fakeHeader{
			"User-Agent":      "probe/1.0",
			"X-Forwarded-For": "203.0.113.7, 10.0.0.1",
		},
	}
	ctx := transport.NewServerContext(context.Background(), tr)

	// Subject-less (no SessionAuth above, if ever wired): attribution
	// columns stay empty.
	if _, err := handler(ctx, nil); err != nil {
		t.Fatal(err)
	}
	ev := sink.collected()[0]
	if ev.Actor != "" || ev.PrincipalType != "" || ev.TenantID != "" {
		t.Errorf("subject-less event carries attribution: %+v", ev)
	}
	if ev.IP != "203.0.113.7" {
		t.Errorf("ip = %q, want first forwarded hop", ev.IP)
	}
	if ev.UserAgent != "probe/1.0" {
		t.Errorf("user agent = %q", ev.UserAgent)
	}

	// With a subject the attribution flows from the context.
	ctx = security.WithSubject(ctx, &security.Subject{UserID: "u1", TenantID: "t1"})
	ctx = transport.NewServerContext(ctx, &fakeTransport{operation: "/cyber.system.v1.UserService/DeleteUser"})
	if _, err := handler(ctx, nil); err != nil {
		t.Fatal(err)
	}
	ev = sink.collected()[1]
	if ev.Actor != "u1" || ev.TenantID != "t1" || ev.PrincipalType != "user" {
		t.Errorf("subject attribution missing: %+v", ev)
	}
}

func TestAuditRequiresSink(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Audit without a sink must panic at assembly")
		}
	}()
	_ = audit.Audit(nil)
}

func TestDenyReasonOfChain(t *testing.T) {
	// Two wrapper layers above the sentinel: identity survives the whole
	// chain, not just the direct rebinding shape.
	wrapped := kerrors.Forbidden("SOMETHING_ELSE", "").WithCause(
		systempb.ErrorSystemAuthzPolicyDenied("").WithCause(kauthz.ErrPolicyDenied),
	)
	if got := audit.DenyReasonOfForTest(wrapped); got != audit.DenyAbacConstraint {
		t.Errorf("deny reason = %q, want %q through two wrapper layers", got, audit.DenyAbacConstraint)
	}
	if got := audit.DenyReasonOfForTest(kerrors.Forbidden("NO_GRANT_X", "")); got != "" {
		t.Errorf("unknown forbidden reason classified as %q", got)
	}
	if got := audit.DenyReasonOfForTest(nil); got != "" {
		t.Errorf("nil error classified as %q", got)
	}
}

func TestActionOfLadder(t *testing.T) {
	if got := audit.ActionOfForTest("/cyber.system.v1.UserService/ListUsers"); got != int32(extv1.Action_ACTION_READ) {
		t.Errorf("annotated read = %d, want READ", got)
	}
	if got := audit.ActionOfForTest("/cyber.system.v1.AuthService/Login"); got != int32(extv1.Action_ACTION_WRITE) {
		t.Errorf("annotated write (login) = %d, want WRITE", got)
	}
	// Unresolvable operation: the WRITE floor, fail-safe.
	if got := audit.ActionOfForTest("/nope.Nope/Nope"); got != int32(extv1.Action_ACTION_WRITE) {
		t.Errorf("floor = %d, want WRITE", got)
	}
}

func TestReadVerb(t *testing.T) {
	for _, m := range []string{"ListUsers", "GetUser"} {
		if !audit.ReadVerbForTest(m) {
			t.Errorf("%s should read as read", m)
		}
	}
	for _, m := range []string{"CreateUser", "RemoveRoleMember", "Login"} {
		if audit.ReadVerbForTest(m) {
			t.Errorf("%s should read as write", m)
		}
	}
}

// The Emitter contract: bounded queue, drop-on-full counted, graceful drain
// on stop.
func TestEmitterDropAndDrain(t *testing.T) {
	var mu sync.Mutex
	var published []*audit.Event
	entered := make(chan struct{})
	gate := make(chan struct{})
	var once sync.Once

	sink, stop := audit.NewEmitter(func(_ context.Context, ev *audit.Event) error {
		once.Do(func() { close(entered) })
		<-gate
		mu.Lock()
		defer mu.Unlock()
		published = append(published, ev)
		return nil
	}, quietLogger(), audit.Capacity(2))

	emit := func(op string) {
		sink.Emit(context.Background(), &audit.Event{Operation: op})
	}
	emit("e1") // taken by the drainer, blocked inside publish
	<-entered
	emit("e2") // queued
	emit("e3") // queued
	emit("e4") // dropped — queue full

	if dropped := sink.Dropped(); dropped != 1 {
		t.Fatalf("dropped = %d, want 1", dropped)
	}
	close(gate)
	stop()

	mu.Lock()
	defer mu.Unlock()
	if len(published) != 3 {
		t.Fatalf("published %d events, want 3 (in-flight + queued)", len(published))
	}
}
