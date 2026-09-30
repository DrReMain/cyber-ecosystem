package authz_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/kratos/security/internal/teststub"
)

type fakeAuthorizer struct {
	decision *authz.Decision
	err      error
	calls    int
	seenOp   string
	seenSubj *security.Subject
}

func (f *fakeAuthorizer) Decide(_ context.Context, s *security.Subject, op string) (*authz.Decision, error) {
	f.calls++
	f.seenOp = op
	f.seenSubj = s
	return f.decision, f.err
}

func TestOperationAuthz(t *testing.T) {
	op := "/cyber.system.v1.UserService/ListUsers"
	subject := &security.Subject{UserID: "u1", TenantID: "default"}
	authedCtx := func() context.Context {
		return security.WithSubject(transport.NewServerContext(context.Background(), teststub.NewTransport(op)), subject)
	}
	var seen *authz.Decision
	var handlerRan bool
	handler := func(ctx context.Context, req any) (any, error) {
		handlerRan = true
		seen, _ = authz.DecisionFromCtx(ctx)
		return "ok", nil
	}

	scoped := &authz.Decision{Allowed: true, Scopes: []authz.ScopeDescriptor{{Entity: "user", Kind: "self"}}}
	builtin := &authz.Decision{Allowed: true, Reason: authz.AllowBuiltin}
	tests := []struct {
		name        string
		ctx         context.Context
		authorizer  *fakeAuthorizer
		wantErr     string
		wantHandler bool
	}{
		{
			name:        "no subject fails closed",
			ctx:         transport.NewServerContext(context.Background(), teststub.NewTransport(op)),
			authorizer:  &fakeAuthorizer{decision: scoped},
			wantErr:     "MISSING_SUBJECT",
			wantHandler: false,
		},
		{
			name:        "no transport fails closed",
			ctx:         security.WithSubject(context.Background(), subject),
			authorizer:  &fakeAuthorizer{decision: scoped},
			wantErr:     "MISSING_SUBJECT",
			wantHandler: false,
		},
		{
			name:        "no grant maps to permission denied",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: &authz.Decision{Reason: authz.DenyNoGrant}},
			wantErr:     "PERMISSION_DENIED",
			wantHandler: false,
		},
		{
			name:        "abac constraint maps to policy denied",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: &authz.Decision{Reason: authz.DenyAbacConstraint}},
			wantErr:     "AUTHZ_POLICY_DENIED",
			wantHandler: false,
		},
		{
			name:        "unknown deny reason maps to permission denied",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: &authz.Decision{Reason: "SOMETHING_ELSE"}},
			wantErr:     "PERMISSION_DENIED",
			wantHandler: false,
		},
		{
			name:        "empty reason deny maps to permission denied",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: &authz.Decision{}},
			wantErr:     "PERMISSION_DENIED",
			wantHandler: false,
		},
		{
			name:        "nil decision fails closed instead of panicking",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{},
			wantErr:     "PERMISSION_DENIED",
			wantHandler: false,
		},
		{
			name:        "facility error propagates",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{err: errors.New("boom")},
			wantErr:     "boom",
			wantHandler: false,
		},
		{
			name:        "allowed subject passes through with scopes",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: scoped},
			wantErr:     "",
			wantHandler: true,
		},
		{
			name:        "builtin allow passes through without scopes",
			ctx:         authedCtx(),
			authorizer:  &fakeAuthorizer{decision: builtin},
			wantErr:     "",
			wantHandler: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen, handlerRan = nil, false
			_, err := authz.OperationAuthz(tt.authorizer)(handler)(tt.ctx, nil)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("err = %v, want nil", err)
			case tt.wantErr != "" && err == nil:
				t.Fatalf("err = nil, want %q", tt.wantErr)
			case tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr):
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
			if handlerRan != tt.wantHandler {
				t.Fatalf("handler ran = %v, want %v", handlerRan, tt.wantHandler)
			}
			if !tt.wantHandler {
				return
			}
			if seen != tt.authorizer.decision {
				t.Fatalf("decision in ctx = %+v, want the authorizer's decision", seen)
			}
			if tt.authorizer.calls != 1 {
				t.Fatalf("authorizer calls = %d, want 1", tt.authorizer.calls)
			}
			if tt.authorizer.seenOp != op {
				t.Fatalf("operation passed to Decide = %q, want %q", tt.authorizer.seenOp, op)
			}
			if tt.authorizer.seenSubj != subject {
				t.Fatalf("subject passed to Decide = %v, want the ctx subject", tt.authorizer.seenSubj)
			}
		})
	}
}

func TestDecisionCtx(t *testing.T) {
	if _, ok := authz.DecisionFromCtx(context.Background()); ok {
		t.Fatal("empty ctx must not carry a decision")
	}
}
