package match

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
	"cyber-ecosystem/app/services/esports/internal/module/test"
)

// The switch-order law under test: the target genre validates before
// anything closes — a switch at an unknown genre returns the domain error
// with the active session and its remaining budget untouched.
func TestSwitchUnknownGenreKeepsActive(t *testing.T) {
	rp := &fakeMatchRP{}
	uc := NewMatchUC(slog.Default(), nil, rp, &fakeMatchTestRP{}, nil, nil)
	_, err := uc.Switch(memberCtx(), "rts")
	if !esportspb.IsEsportsMatchGenreUnknown(err) {
		t.Fatalf("err = %v, want ESPORTS_MATCH_GENRE_UNKNOWN", err)
	}
	if len(rp.statusCalls) != 0 {
		t.Fatalf("status calls = %v, want none (nothing closed)", rp.statusCalls)
	}
}

// The happy-path switch under test: the active session closes as switched
// and the replacement is born for the target genre.
func TestSwitchClosesThenCreates(t *testing.T) {
	rp := &fakeMatchRP{
		configs: []MatchConfig{{ID: "c1", Genre: "fps", Ability: "reactionSpeed", RenderOrder: 1}},
		active:  &MatchSession{ID: "s1", UserID: "u1", Genre: "moba", Status: statusActive, CreatedAt: time.Now()},
	}
	uc := NewMatchUC(slog.Default(), nil, rp, &fakeMatchTestRP{tests: []*test.Test{
		{Code: "red_dot", Ability: utils.Ptr("reactionSpeed"), Category: utils.Ptr("fps")},
	}}, nil, nil)
	view, err := uc.Switch(memberCtx(), "fps")
	if err != nil {
		t.Fatal(err)
	}
	if len(rp.statusCalls) != 1 || rp.statusCalls[0] != "s1:"+statusSwitched {
		t.Fatalf("status calls = %v, want s1 switched", rp.statusCalls)
	}
	if !view.Created || view.Session.Genre != "fps" || view.Session.Status != statusActive {
		t.Fatalf("view = %+v, want a freshly created fps session", view)
	}
}

// The create-race law under test: two concurrent create-or-resume reads
// both miss the active row and both draw — the partial unique lets one
// insert win, and the loser re-reads and returns the survivor as a fresh
// birth instead of surfacing the constraint.
func TestGetCreateRaceReturnsSurvivor(t *testing.T) {
	survivor := &MatchSession{
		ID: "s2", UserID: "u1", Genre: "fps", Difficulty: "experience", Seed: 7,
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
		RetestBudget: 3, Status: statusActive, CreatedAt: time.Now(),
	}
	reads := 0
	rp := &fakeMatchRP{
		configs:   []MatchConfig{{ID: "c1", Genre: "fps", Ability: "reactionSpeed", RenderOrder: 1}},
		createErr: errorspb.ErrorInfraErrorDbConstraint(""),
		findActive: func() (*MatchSession, error) {
			reads++
			if reads == 1 {
				return nil, nil
			}
			return survivor, nil
		},
	}
	uc := NewMatchUC(slog.Default(), nil, rp, &fakeMatchTestRP{tests: []*test.Test{
		{Code: "red_dot", Ability: utils.Ptr("reactionSpeed"), Category: utils.Ptr("fps")},
	}}, nil, nil)
	view, err := uc.Get(memberCtx(), utils.Ptr("fps"))
	if err != nil {
		t.Fatal(err)
	}
	if view.Session.ID != "s2" || !view.Created {
		t.Fatalf("view = %+v, want the surviving session s2 reported as created", view)
	}
}

// The peek law under test: a pure read — an in-window session reports, an
// over-window one reads absent (without Get's lazy expire write), and no
// active row reads absent; none of the three paths creates or writes.
func TestPeekIsReadOnly(t *testing.T) {
	rp := &fakeMatchRP{active: &MatchSession{
		ID: "s1", UserID: "u1", Genre: "fps", Status: statusActive, CreatedAt: time.Now(),
		Steps:        []types.MatchStep{{Ability: "reactionSpeed", TestCode: "red_dot"}},
		RetestBudget: 1,
	}}
	uc := NewMatchUC(slog.Default(), nil, rp, &fakeMatchTestRP{}, nil, nil)
	view, err := uc.Peek(memberCtx())
	if err != nil {
		t.Fatal(err)
	}
	if view == nil || view.Session.ID != "s1" {
		t.Fatalf("view = %+v, want the in-window session s1", view)
	}

	// Over-window: absent, and the lazy expire stays Get's job.
	rp.active.CreatedAt = time.Now().Add(-matchResumeWindow - time.Hour)
	view, err = uc.Peek(memberCtx())
	if err != nil {
		t.Fatal(err)
	}
	if view != nil {
		t.Fatalf("over-window view = %+v, want absent", view)
	}

	// No active row: absent.
	rp.active = nil
	view, err = uc.Peek(memberCtx())
	if err != nil {
		t.Fatal(err)
	}
	if view != nil {
		t.Fatalf("no-active view = %+v, want absent", view)
	}

	if len(rp.statusCalls) != 0 || rp.created != nil {
		t.Fatalf("statusCalls = %v, created = %+v — peek must not write", rp.statusCalls, rp.created)
	}
}

func memberCtx() context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "u1"})
}

type fakeMatchRP struct {
	MatchRP
	configs     []MatchConfig
	active      *MatchSession
	createErr   error
	statusCalls []string
	created     *MatchSession
	findActive  func() (*MatchSession, error)
}

func (rp *fakeMatchRP) ListConfig(_ context.Context, _ string) ([]MatchConfig, error) {
	return rp.configs, nil
}

func (rp *fakeMatchRP) FindActive(_ context.Context, _ string) (*MatchSession, error) {
	if rp.findActive != nil {
		return rp.findActive()
	}
	return rp.active, nil
}

func (rp *fakeMatchRP) UpdateStatus(_ context.Context, id, status string) error {
	rp.statusCalls = append(rp.statusCalls, id+":"+status)
	return nil
}

func (rp *fakeMatchRP) Create(_ context.Context, s *MatchSession) (*MatchSession, error) {
	if rp.createErr != nil {
		return nil, rp.createErr
	}
	rp.created = s
	return s, nil
}

func (rp *fakeMatchRP) CountByUser(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (rp *fakeMatchRP) CountByUserGenre(_ context.Context, _ string, _ string) (int, error) {
	return 0, nil
}

// The catalog read the UC needs beyond its own aggregate: the pool feed and
// the derived done/retest counts.
type fakeMatchTestRP struct {
	test.TestRP
	tests []*test.Test
}

func (rp *fakeMatchTestRP) List(_ context.Context) ([]*test.Test, error) {
	return rp.tests, nil
}

func (rp *fakeMatchTestRP) CountBySessionCodes(_ context.Context, _ string) (map[string]int, error) {
	return map[string]int{}, nil
}
