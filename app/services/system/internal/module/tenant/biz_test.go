package tenant

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// passthroughTm runs the unit inline.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeTenantRP answers the registry surface with in-memory rows and records
// the write side.
type fakeTenantRP struct {
	TenantRP
	rows    map[string]*Tenant
	findErr error

	listOut *TenantListOut

	updatedIDs  []string
	updatedVals []bool
}

func (f *fakeTenantRP) FindByID(_ context.Context, id string) (*Tenant, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeTenantRP) FindByCode(_ context.Context, code string) (*Tenant, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	for _, r := range f.rows {
		if r.Code == code {
			return r, nil
		}
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeTenantRP) List(context.Context, *TenantListIn) (*TenantListOut, error) {
	return f.listOut, nil
}

func (f *fakeTenantRP) UpdateStatus(_ context.Context, id string, enabled bool) (*Tenant, error) {
	f.updatedIDs = append(f.updatedIDs, id)
	f.updatedVals = append(f.updatedVals, enabled)
	r := f.rows[id]
	r.Enabled = enabled
	return r, nil
}

func newTenantUC(rp *fakeTenantRP) *TenantUC {
	return NewTenantUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp,
	)
}

func defaultRow(enabled bool) *Tenant {
	return &Tenant{ID: "t-default", Code: security.DefaultTenant, Name: "Default", Enabled: enabled}
}

// Get ------------------------------------------------------------------------------------------------------------------

// An unknown id classifies as TENANT_NOT_FOUND; a generic store error keeps
// its infra identity.
func TestGetClassifiedGenericPassthrough(t *testing.T) {
	uc := newTenantUC(&fakeTenantRP{rows: map[string]*Tenant{}})

	if _, err := uc.Get(context.Background(), "ghost"); !systempb.IsSystemTenantNotFound(err) {
		t.Fatalf("unknown: want TENANT_NOT_FOUND, got %v", err)
	}

	uc2 := newTenantUC(&fakeTenantRP{findErr: errorspb.ErrorInfraErrorNetworkConnection("")})
	if _, err := uc2.Get(context.Background(), "t1"); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic: want passthrough, got %v", err)
	}
}

// A known id returns its row.
func TestGetFound(t *testing.T) {
	uc := newTenantUC(&fakeTenantRP{rows: map[string]*Tenant{"t1": {ID: "t1", Code: "acme", Name: "Acme", Enabled: true}}})

	out, err := uc.Get(context.Background(), "t1")
	if err != nil || out.Code != "acme" {
		t.Fatalf("get: out=%+v err=%v", out, err)
	}
}

// List -----------------------------------------------------------------------------------------------------------------

// List is a pure delegation — filters and paging live in the repo.
func TestListDelegates(t *testing.T) {
	want := &TenantListOut{List: []*Tenant{defaultRow(true)}}
	uc := newTenantUC(&fakeTenantRP{listOut: want})

	out, err := uc.List(context.Background(), &TenantListIn{})
	if err != nil || out != want {
		t.Fatalf("list: out=%v err=%v", out, err)
	}
}

// SetEnabled -----------------------------------------------------------------------------------------------------------

// The default tenant is the degenerate mode's whole world: disabling it
// bricks every login including the admin who could undo it. The registry
// refuses at the only write entrance.
func TestSetEnabledDefaultDisableProtected(t *testing.T) {
	rp := &fakeTenantRP{rows: map[string]*Tenant{"t-default": defaultRow(true)}}
	uc := newTenantUC(rp)

	if _, err := uc.UpdateStatus(context.Background(), "t-default", false); !systempb.IsSystemTenantProtected(err) {
		t.Fatalf("default disable: want TENANT_PROTECTED, got %v", err)
	}
	if len(rp.updatedIDs) != 0 {
		t.Fatal("default disable: no write may land")
	}
}

// Re-enabling the default row stays open — it is the recovery path when the
// row was disabled out-of-band.
func TestSetEnabledDefaultReenableAllowed(t *testing.T) {
	rp := &fakeTenantRP{rows: map[string]*Tenant{"t-default": defaultRow(false)}}
	uc := newTenantUC(rp)

	if _, err := uc.UpdateStatus(context.Background(), "t-default", true); err != nil {
		t.Fatalf("default re-enable: %v", err)
	}
	if len(rp.updatedIDs) != 1 || !rp.updatedVals[0] {
		t.Fatalf("default re-enable: writes = %v %v", rp.updatedIDs, rp.updatedVals)
	}
}

// A status write that changes nothing writes nothing.
func TestSetEnabledCosmeticSkip(t *testing.T) {
	rp := &fakeTenantRP{rows: map[string]*Tenant{"t1": {ID: "t1", Code: "acme", Enabled: true}}}
	uc := newTenantUC(rp)

	out, err := uc.UpdateStatus(context.Background(), "t1", true)
	if err != nil || !out.Enabled {
		t.Fatalf("cosmetic: out=%+v err=%v", out, err)
	}
	if len(rp.updatedIDs) != 0 {
		t.Fatal("cosmetic: zero writes expected")
	}
}

// An unknown id answers TENANT_NOT_FOUND with no write.
func TestSetEnabledUnknownRefuses(t *testing.T) {
	rp := &fakeTenantRP{rows: map[string]*Tenant{}}
	uc := newTenantUC(rp)

	if _, err := uc.UpdateStatus(context.Background(), "ghost", false); !systempb.IsSystemTenantNotFound(err) {
		t.Fatalf("unknown: want TENANT_NOT_FOUND, got %v", err)
	}
	if len(rp.updatedIDs) != 0 {
		t.Fatal("unknown: no write may land")
	}
}

// The happy path flips the flag through the repo and answers the updated
// row.
func TestSetEnabledHappy(t *testing.T) {
	rp := &fakeTenantRP{rows: map[string]*Tenant{"t1": {ID: "t1", Code: "acme", Enabled: true}}}
	uc := newTenantUC(rp)

	out, err := uc.UpdateStatus(context.Background(), "t1", false)
	if err != nil || out.Enabled {
		t.Fatalf("disable: out=%+v err=%v", out, err)
	}
	if len(rp.updatedIDs) != 1 || rp.updatedVals[0] {
		t.Fatalf("disable: writes = %v %v", rp.updatedIDs, rp.updatedVals)
	}
}
