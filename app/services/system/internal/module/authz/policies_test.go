package authz

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
)

// White-box anchors for the frozen evaluation kernel: the lazy
// zero-resolution path, AND semantics without short-circuit (explain wants
// every policy's state), the two fail-closed rows (unregistered kind,
// evaluation error), and order-preserving filtration. The
// resolveAttributes-error path is a one-line wrap around source resolution
// and needs a polluting global registration — its fail-closed law stays
// pinned by the e2e face and the conventions guards, not here.

var (
	subj    = &security.Subject{UserID: "u1"}
	discard = slog.New(slog.DiscardHandler)
	// frozenEval pins the evaluation clock: survivingGrants runs the real
	// timeSource seam, and the all-day/NONE fixtures below hold at any
	// instant, so the freeze makes that independence explicit.
	frozenEval = timeSource{clock: func() time.Time { return time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local) }}
)

func allDayWindow() PolicyView {
	return PolicyView{
		Kind: kauthz.PolicyKindTimeWindow,
		Name: "all-day",
		Params: map[string]any{
			"windows": []any{map[string]any{"start": "00:00", "end": "24:00"}},
		},
	}
}

func neverCalendar() PolicyView {
	return PolicyView{
		Kind:   kauthz.PolicyKindCalendar,
		Name:   "never",
		Params: map[string]any{"base": kauthz.CalendarBaseNone},
	}
}

// Law: no linked policy → the grants pass through untouched, zero attribute
// resolution (the lazy path is load-bearing: no policy, no dialing sources).
func TestSurvivingGrantsLazyPassthrough(t *testing.T) {
	grants := []GrantView{{Pattern: "/a/*"}, {Pattern: "/b/Op", ScopeKind: "self"}}
	out := survivingGrants(context.Background(), discard, subj, grants, []kauthz.AttributeSource{frozenEval})
	if len(out) != 2 || out[0].Pattern != "/a/*" || out[1].Pattern != "/b/Op" {
		t.Fatalf("policy-free grants must pass through untouched, got %+v", out)
	}
}

// Law: surviving = order-preserving filtration by the grant's AND — the
// all-day window survives, the NONE calendar does not, the untouched
// neighbor keeps its slot.
func TestSurvivingGrantsFiltersInOrder(t *testing.T) {
	grants := []GrantView{
		{Pattern: "/pass/*", Policies: []PolicyView{allDayWindow()}},
		{Pattern: "/bare/Op"},
		{Pattern: "/fail/*", Policies: []PolicyView{neverCalendar()}},
	}
	out := survivingGrants(context.Background(), discard, subj, grants, []kauthz.AttributeSource{frozenEval})
	if len(out) != 2 || out[0].Pattern != "/pass/*" || out[1].Pattern != "/bare/Op" {
		t.Fatalf("filtration must keep pass+bare in order, got %+v", out)
	}
}

// Law: AND semantics without short-circuit — one failing policy eliminates
// the grant, and every policy still reports its state (explain's contract).
func TestEvalGrantPoliciesANDNoShortCircuit(t *testing.T) {
	attrs := map[string]any{kauthz.AttrTimeNow: time.Date(2026, 3, 10, 12, 0, 0, 0, time.Local)}
	passed, states := evalGrantPolicies(context.Background(), discard, attrs,
		[]PolicyView{allDayWindow(), neverCalendar()})
	if passed {
		t.Fatal("one failed policy must eliminate the grant")
	}
	if len(states) != 2 {
		t.Fatalf("no short-circuit: both states must report, got %d", len(states))
	}
	if states[0].State != policyPassed || states[1].State != policyFailed {
		t.Fatalf("states must be [passed failed], got %+v", states)
	}
}

// Law: the two fail-closed rows — an unregistered kind and an evaluation
// error both mark the policy failed, never skipped and never an allow.
func TestEvalGrantPoliciesFailsClosed(t *testing.T) {
	attrs := map[string]any{kauthz.AttrTimeNow: time.Now()}

	passed, states := evalGrantPolicies(context.Background(), discard, attrs,
		[]PolicyView{{Kind: "future_kind", Name: "ghost", Params: map[string]any{}}})
	if passed || states[0].State != policyFailed {
		t.Fatalf("unregistered kind must fail closed, got passed=%v state=%+v", passed, states[0])
	}

	broken := PolicyView{
		Kind: kauthz.PolicyKindTimeWindow, Name: "inverted",
		Params: map[string]any{"windows": []any{map[string]any{"start": "17:00", "end": "09:00"}}},
	}
	passed, states = evalGrantPolicies(context.Background(), discard, attrs, []PolicyView{broken})
	if passed || states[0].State != policyFailed {
		t.Fatalf("evaluation error must fail closed, got passed=%v state=%+v", passed, states[0])
	}
}

// Law: neededAttrKeys unions only registered plugins' keys — an unknown
// kind contributes nothing here (its failure is evaluation's business).
func TestNeededAttrKeysSkipUnknownKinds(t *testing.T) {
	needed := neededAttrKeys([]GrantView{
		{Pattern: "/x/*", Policies: []PolicyView{{Kind: "future_kind", Params: map[string]any{}}}},
	})
	if len(needed) != 0 {
		t.Fatalf("unknown kinds contribute no attribute keys, got %v", needed)
	}
	needed = neededAttrKeys([]GrantView{
		{Pattern: "/x/*", Policies: []PolicyView{allDayWindow()}},
	})
	if _, ok := needed[kauthz.AttrTimeNow]; !ok {
		t.Fatalf("a time-window grant must need the time attribute, got %v", needed)
	}
}
