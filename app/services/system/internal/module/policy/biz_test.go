package policy

import (
	"context"
	"io"
	"log/slog"
	"testing"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// passthroughTm runs the unit inline — these laws are about validation and
// broadcast conditions, not transaction semantics.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakePolicyRP counts the write side; rows back the FindByID checks.
type fakePolicyRP struct {
	PolicyRP
	rows    map[string]*Policy
	creates int
	updates int
	deletes []string
	notify  int
}

func (f *fakePolicyRP) FindByID(_ context.Context, id string) (*Policy, error) {
	if p, ok := f.rows[id]; ok {
		return p, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakePolicyRP) Create(_ context.Context, p *Policy) (*Policy, error) {
	f.creates++
	out := *p
	out.ID = "p-new"
	return &out, nil
}

func (f *fakePolicyRP) Update(_ context.Context, _ []string, p *Policy) (*Policy, error) {
	f.updates++
	return p, nil
}

func (f *fakePolicyRP) Delete(_ context.Context, id string) (string, error) {
	f.deletes = append(f.deletes, id)
	return id, nil
}

func (f *fakePolicyRP) NotifyChanged(context.Context) { f.notify++ }

func newPolicyUC(rp *fakePolicyRP) *PolicyUC {
	return NewPolicyUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp,
	)
}

func validWindows() map[string]any {
	return map[string]any{"windows": []any{map[string]any{"start": "09:00", "end": "18:00"}}}
}

// Create --------------------------------------------------------------------------------------------------------------

// A create with invalid params is rejected before the row exists.
func TestCreateValidatesParams(t *testing.T) {
	rp := &fakePolicyRP{}
	uc := newPolicyUC(rp)

	p := &Policy{Kind: kauthz.PolicyKindTimeWindow, Params: map[string]any{"windows": []any{}}}
	if _, err := uc.Create(context.Background(), p); !systempb.IsSystemPolicyInvalidParams(err) {
		t.Fatalf("invalid params: want POLICY_INVALID_PARAMS, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("invalid params: no row may be born")
	}
}

// A valid create broadcasts the rebuild — a new policy can constrain live
// verdicts on the next compile.
func TestCreateHappyNotifies(t *testing.T) {
	rp := &fakePolicyRP{}
	uc := newPolicyUC(rp)

	p := &Policy{Kind: kauthz.PolicyKindTimeWindow, Params: validWindows()}
	if _, err := uc.Create(context.Background(), p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if rp.creates != 1 || rp.notify != 1 {
		t.Fatalf("create: creates=%d notify=%d", rp.creates, rp.notify)
	}
}

// Update --------------------------------------------------------------------------------------------------------------

// An unknown policy classifies as POLICY_NOT_FOUND.
func TestUpdateUnknown(t *testing.T) {
	uc := newPolicyUC(&fakePolicyRP{rows: map[string]*Policy{}})

	if _, err := uc.Update(context.Background(), []string{"name"}, &Policy{ID: "ghost"}); !systempb.IsSystemPolicyNotFound(err) {
		t.Fatalf("unknown: want POLICY_NOT_FOUND, got %v", err)
	}
}

// The kind derives from the oneof case — reconfiguring a row into a different
// policy type in place is refused; a different kind is a different row.
func TestUpdateKindImmutable(t *testing.T) {
	rp := &fakePolicyRP{rows: map[string]*Policy{"p1": {ID: "p1", Kind: kauthz.PolicyKindTimeWindow, Params: validWindows()}}}
	uc := newPolicyUC(rp)

	p := &Policy{ID: "p1", Kind: kauthz.PolicyKindCalendar, Params: map[string]any{"base": kauthz.CalendarBaseAll}}
	if _, err := uc.Update(context.Background(), []string{"params"}, p); !systempb.IsSystemPolicyInvalidParams(err) {
		t.Fatalf("kind flip: want POLICY_INVALID_PARAMS, got %v", err)
	}
	if rp.updates != 0 {
		t.Fatal("kind flip: no write may happen")
	}
}

// Params and enabled flips reach the snapshot; a name-only edit never does —
// cosmetic edits skip the rebuild broadcast.
func TestUpdateCosmeticSkipsNotify(t *testing.T) {
	rp := &fakePolicyRP{rows: map[string]*Policy{"p1": {ID: "p1", Kind: kauthz.PolicyKindTimeWindow, Params: validWindows()}}}
	uc := newPolicyUC(rp)

	if _, err := uc.Update(context.Background(), []string{"name"}, &Policy{ID: "p1", Name: "n"}); err != nil {
		t.Fatalf("cosmetic: %v", err)
	}
	if rp.notify != 0 {
		t.Fatalf("cosmetic: notify = %d, want 0", rp.notify)
	}

	if _, err := uc.Update(context.Background(), []string{"params"}, &Policy{ID: "p1", Kind: kauthz.PolicyKindTimeWindow, Params: validWindows()}); err != nil {
		t.Fatalf("params update: %v", err)
	}
	if rp.notify != 1 {
		t.Fatalf("params update: notify = %d, want 1", rp.notify)
	}
}

// Deleting unlinks the policy from its grants — a relax-direction change
// replicas must see.
func TestDeleteNotifies(t *testing.T) {
	rp := &fakePolicyRP{rows: map[string]*Policy{"p1": {ID: "p1"}}}
	uc := newPolicyUC(rp)

	if _, err := uc.Delete(context.Background(), "p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deletes) != 1 || rp.notify != 1 {
		t.Fatalf("delete: deletes=%v notify=%d", rp.deletes, rp.notify)
	}
}

// An unknown id on the read face classifies; a generic store error keeps its
// infra identity.
func TestGetClassifiedGenericPassthrough(t *testing.T) {
	uc := newPolicyUC(&fakePolicyRP{rows: map[string]*Policy{}})

	if _, err := uc.Get(context.Background(), "ghost"); !systempb.IsSystemPolicyNotFound(err) {
		t.Fatalf("unknown: want POLICY_NOT_FOUND, got %v", err)
	}
}

// validateParams ------------------------------------------------------------------------------------------------------

// Time windows are same-day spans: canonical HH:MM, "24:00" ends the day,
// start < end — cross-midnight is two windows, not one.
func TestValidateTimeWindow(t *testing.T) {
	cases := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{"valid", validWindows(), false},
		{"empty windows", map[string]any{"windows": []any{}}, true},
		{"missing start", map[string]any{"windows": []any{map[string]any{"end": "18:00"}}}, true},
		{"bad hh:mm", map[string]any{"windows": []any{map[string]any{"start": "9:00", "end": "18:00"}}}, true},
		{"end 24:00 ok", map[string]any{"windows": []any{map[string]any{"start": "09:00", "end": "24:00"}}}, false},
		{"start after end", map[string]any{"windows": []any{map[string]any{"start": "18:00", "end": "09:00"}}}, true},
		{"span not object", map[string]any{"windows": []any{"x"}}, true},
	}
	for _, c := range cases {
		err := validateParams(kauthz.PolicyKindTimeWindow, c.params)
		if got := err != nil; got != c.wantErr {
			t.Fatalf("%s: err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

// Calendar params: a closed base vocabulary, real dates only, one override
// per date.
func TestValidateCalendar(t *testing.T) {
	cases := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{"valid", map[string]any{"base": kauthz.CalendarBaseWeekdays}, false},
		{"unknown base", map[string]any{"base": "SOMETIMES"}, true},
		{
			"valid override",
			map[string]any{"base": kauthz.CalendarBaseAll, "overrides": []any{map[string]any{"date": "2026-10-01"}}},
			false,
		},
		{
			"invalid date",
			map[string]any{"base": kauthz.CalendarBaseAll, "overrides": []any{map[string]any{"date": "2026-13-01"}}},
			true,
		},
		{
			"duplicate date",
			map[string]any{"base": kauthz.CalendarBaseAll, "overrides": []any{
				map[string]any{"date": "2026-10-01"},
				map[string]any{"date": "2026-10-01"},
			}},
			true,
		},
		{"override not object", map[string]any{"base": kauthz.CalendarBaseAll, "overrides": []any{"x"}}, true},
	}
	for _, c := range cases {
		err := validateParams(kauthz.PolicyKindCalendar, c.params)
		if got := err != nil; got != c.wantErr {
			t.Fatalf("%s: err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

// The unknown-kind guard is for writes that bypass the API's oneof — it fails
// closed as POLICY_UNKNOWN_KIND.
func TestValidateParamsUnknownKind(t *testing.T) {
	if err := validateParams("no-such-kind", nil); !systempb.IsSystemPolicyUnknownKind(err) {
		t.Fatalf("unknown kind: want POLICY_UNKNOWN_KIND, got %v", err)
	}
}
