package authz

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

type fakeAuthzRP struct {
	decision *kauthz.Decision
	err      error
	calls    int
}

func (f *fakeAuthzRP) CheckGrants(ctx context.Context, subject *security.Subject, operation string) (*kauthz.Decision, error) {
	f.calls++
	return f.decision, f.err
}

func TestDecide(t *testing.T) {
	cases := []struct {
		name        string
		rpDecision  *kauthz.Decision
		rpErr       error
		wantAllowed bool
		wantErr     bool
	}{
		{
			name:        "forwards the engine verdict on allow",
			rpDecision:  &kauthz.Decision{Allowed: true},
			wantAllowed: true,
		},
		{
			name:       "deny verdict passes through unmangled",
			rpDecision: &kauthz.Decision{Allowed: false, Reason: "no surviving grant"},
		},
		{
			name:    "fail-closed: engine error surfaces, never an allow",
			rpErr:   errors.New("engine unavailable"),
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rp := &fakeAuthzRP{decision: tc.rpDecision, err: tc.rpErr}
			uc := NewAuthzUC(slog.New(slog.DiscardHandler), nil, rp)
			decision, err := uc.Decide(context.Background(), &security.Subject{UserID: "u1"}, "agentpb.AgentChat")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected the engine error to surface, got decision %+v", decision)
				}
				if decision != nil && decision.Allowed {
					t.Fatal("an errored check must never come back as an allow")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decision == nil {
				t.Fatal("verdict must never be nil on the success path")
			}
			if decision.Allowed != tc.wantAllowed {
				t.Fatalf("verdict mangled in transit: engine said %+v, Decide returned %+v", tc.rpDecision, decision)
			}
			if rp.calls != 1 {
				t.Fatalf("expected exactly one engine call, got %d", rp.calls)
			}
		})
	}
}
