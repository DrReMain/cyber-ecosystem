package population

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// fakePopulationRP answers the registry slice: rows by code, captured
// creates, injectable failures for the guard and race laws.
type fakePopulationRP struct {
	PopulationRP
	rows      map[string]*Population
	findErr   error
	creates   []*Population
	createErr error
	listOut   *PopulationListOut
	notifies  int
}

func (f *fakePopulationRP) NotifyChanged(context.Context) {
	f.notifies++
}

func (f *fakePopulationRP) FindByCode(_ context.Context, code string) (*Population, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if p, ok := f.rows[code]; ok {
		return p, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakePopulationRP) Create(_ context.Context, p *Population) (*Population, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.creates = append(f.creates, p)
	return p, nil
}

func (f *fakePopulationRP) List(context.Context, *PopulationListIn) (*PopulationListOut, error) {
	return f.listOut, nil
}

func newPopulationUC(rp *fakePopulationRP) *PopulationUC {
	return NewPopulationUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, rp,
	)
}

// Create --------------------------------------------------------------------------------------------------------------

// A registry row is born exactly as asked — code, name, and archetype
// verbatim; the join key and the vocabulary are the whole payload.
func TestCreateBirthsRegistryRow(t *testing.T) {
	rp := &fakePopulationRP{rows: map[string]*Population{}}
	uc := newPopulationUC(rp)

	p, err := uc.Create(context.Background(), "guest", "Guest", security.ArchetypePatron)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.Code != "guest" || p.Name != "Guest" || p.Archetype != security.ArchetypePatron {
		t.Fatalf("create: row = %+v", p)
	}
	if len(rp.creates) != 1 {
		t.Fatalf("create: writes = %d, want 1", len(rp.creates))
	}
	if rp.notifies != 1 {
		t.Fatalf("birth: notifies = %d, want the engine told once", rp.notifies)
	}
}

// The default population is seed-born; an operator-borne row (a patron one,
// say) would silently turn every default seat birth into a grantless
// identity — reserved rejects even where the seed row is missing.
func TestCreateReservedCodeRejected(t *testing.T) {
	rp := &fakePopulationRP{rows: map[string]*Population{}}
	uc := newPopulationUC(rp)

	if _, err := uc.Create(context.Background(), security.DefaultPopulation, "X", security.ArchetypePatron); !systempb.IsSystemPopulationProtected(err) {
		t.Fatalf("reserved: err = %v, want PROTECTED", err)
	}
	if len(rp.creates) != 0 {
		t.Fatal("reserved: the write must not happen")
	}
}

// The archetype vocabulary is compile-time; a word outside it is a client
// error before any write.
func TestCreateUnknownArchetypeRejected(t *testing.T) {
	rp := &fakePopulationRP{rows: map[string]*Population{}}
	uc := newPopulationUC(rp)

	if _, err := uc.Create(context.Background(), "robot", "Robot", "robot"); !systempb.IsSystemPopulationInvalidArchetype(err) {
		t.Fatalf("archetype: err = %v, want INVALID_ARCHETYPE", err)
	}
	if len(rp.creates) != 0 {
		t.Fatal("archetype: the write must not happen")
	}
}

// A duplicate code is the same shape whether the pre-check catches it or
// the unique index does — the race backstop never leaks an infra error.
func TestCreateDuplicateRejected(t *testing.T) {
	rp := &fakePopulationRP{rows: map[string]*Population{
		"guest": {Code: "guest", Archetype: security.ArchetypePatron},
	}}
	uc := newPopulationUC(rp)
	if _, err := uc.Create(context.Background(), "guest", "Guest", security.ArchetypePatron); !systempb.IsSystemPopulationExists(err) {
		t.Fatalf("duplicate pre-check: err = %v, want EXISTS", err)
	}
	if len(rp.creates) != 0 {
		t.Fatal("duplicate pre-check: the write must not happen")
	}

	race := &fakePopulationRP{rows: map[string]*Population{}, createErr: errorspb.ErrorInfraErrorDbConstraint("")}
	if _, err := newPopulationUC(race).Create(context.Background(), "guest", "Guest", security.ArchetypePatron); !systempb.IsSystemPopulationExists(err) {
		t.Fatalf("duplicate race: err = %v, want EXISTS", err)
	}
}

// A registry read that fails any way other than not-found is infra — it
// surfaces, never reads as "the code is free".
func TestCreateLookupFailureSurfaces(t *testing.T) {
	rp := &fakePopulationRP{rows: map[string]*Population{}, findErr: errorspb.ErrorInfraErrorDbInternal("")}
	uc := newPopulationUC(rp)

	if _, err := uc.Create(context.Background(), "guest", "Guest", security.ArchetypePatron); !errorspb.IsInfraErrorDbInternal(err) {
		t.Fatalf("lookup failure: err = %v, want passthrough", err)
	}
	if len(rp.creates) != 0 {
		t.Fatal("lookup failure: the write must not happen")
	}
}

// List delegates — the registry carries no list-side law of its own.
func TestListDelegates(t *testing.T) {
	want := &PopulationListOut{List: []*Population{{Code: "user"}}}
	rp := &fakePopulationRP{listOut: want}
	got, err := newPopulationUC(rp).List(context.Background(), &PopulationListIn{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got != want {
		t.Fatalf("list: out = %v, want passthrough", got)
	}
}
