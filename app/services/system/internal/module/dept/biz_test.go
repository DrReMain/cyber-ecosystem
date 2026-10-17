package dept

import (
	"context"
	"io"
	"log/slog"
	"testing"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// passthroughTm runs the unit inline — these laws are about guards and write
// order, not transaction semantics.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeDeptRP answers the tree checks (children/users/ancestor) with
// controllable verdicts and counts the write side.
type fakeDeptRP struct {
	DeptRP
	rows           map[string]*Dept
	creates        int
	updates        int
	deletes        []string
	hasChildren    bool
	hasUsers       bool
	isAncestor     bool
	ancestorCalled int
	removed        bool
	membersOut     *MemberListOut
	membersCalls   int
}

func (f *fakeDeptRP) Get(_ context.Context, id string) (*Dept, error) {
	if d, ok := f.rows[id]; ok {
		return d, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeDeptRP) Create(_ context.Context, d *Dept) (*Dept, error) {
	f.creates++
	out := *d
	out.ID = "d-new"
	return &out, nil
}

func (f *fakeDeptRP) Update(_ context.Context, _ []string, d *Dept) (*Dept, error) {
	f.updates++
	return d, nil
}

func (f *fakeDeptRP) Delete(_ context.Context, id string) (string, error) {
	f.deletes = append(f.deletes, id)
	return id, nil
}

func (f *fakeDeptRP) HasChildren(context.Context, string) (bool, error) {
	return f.hasChildren, nil
}

func (f *fakeDeptRP) HasUsers(context.Context, string) (bool, error) {
	return f.hasUsers, nil
}

func (f *fakeDeptRP) IsAncestor(context.Context, string, string) (bool, error) {
	f.ancestorCalled++
	return f.isAncestor, nil
}

func (f *fakeDeptRP) RemoveMember(_ context.Context, _, _ string) (bool, error) {
	return f.removed, nil
}

func (f *fakeDeptRP) ListMembers(_ context.Context, _ *MemberListIn) (*MemberListOut, error) {
	f.membersCalls++
	return f.membersOut, nil
}

func (f *fakeDeptRP) List(_ context.Context, _ *DeptListIn) (*DeptListOut, error) {
	return &DeptListOut{}, nil
}

func newDeptUC(rp *fakeDeptRP) *DeptUC {
	return NewDeptUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp,
	)
}

func strPtr(s string) *string { return &s }

// Create --------------------------------------------------------------------------------------------------------------

// A named parent must exist before a child is born; an unknown parent passes
// its infra identity through — no domain shape is invented.
func TestCreateParentMustExist(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{}}
	uc := newDeptUC(rp)

	if _, err := uc.Create(context.Background(), &Dept{Name: strPtr("x"), ParentID: strPtr("ghost")}); !errorspb.IsInfraErrorDbNotFound(err) {
		t.Fatalf("unknown parent: want infra passthrough, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unknown parent: no row may be created")
	}

	rp2 := &fakeDeptRP{rows: map[string]*Dept{"p1": {ID: "p1"}}}
	if _, err := newDeptUC(rp2).Create(context.Background(), &Dept{ParentID: strPtr("p1")}); err != nil {
		t.Fatalf("known parent: %v", err)
	}
	if rp2.creates != 1 {
		t.Fatal("known parent: the row must be born")
	}
}

// Update: the tree stays acyclic ------------------------------------------------------------

// A dept cannot be its own parent.
func TestUpdateSelfParentRejected(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}}
	uc := newDeptUC(rp)

	if _, err := uc.Update(context.Background(), []string{"parent_id"}, &Dept{ID: "d1", ParentID: strPtr("d1")}); !systempb.IsSystemDeptSelfParent(err) {
		t.Fatalf("self parent: want DEPT_SELF_PARENT, got %v", err)
	}
	if rp.updates != 0 {
		t.Fatal("self parent: no write may happen")
	}
}

// Reparenting under a descendant would forge a cycle — the ancestor check
// rejects it before the write.
func TestUpdateCycleRejected(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}, "d2": {ID: "d2"}}, isAncestor: true}
	uc := newDeptUC(rp)

	if _, err := uc.Update(context.Background(), []string{"parent_id"}, &Dept{ID: "d1", ParentID: strPtr("d2")}); !systempb.IsSystemDeptCycle(err) {
		t.Fatalf("cycle: want DEPT_CYCLE, got %v", err)
	}
	if rp.updates != 0 {
		t.Fatal("cycle: no write may happen")
	}
}

// Clearing to root carries no cycle risk by definition — the ancestor check
// is not even consulted.
func TestUpdateClearToRootSkipsCycleCheck(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}}
	uc := newDeptUC(rp)

	if _, err := uc.Update(context.Background(), []string{"parent_id"}, &Dept{ID: "d1", ParentID: nil}); err != nil {
		t.Fatalf("clear to root: %v", err)
	}
	if rp.ancestorCalled != 0 || rp.updates != 1 {
		t.Fatalf("clear to root: ancestorCalled=%d updates=%d", rp.ancestorCalled, rp.updates)
	}
}

// Delete --------------------------------------------------------------------------------------------------------------

// A parent with children refuses deletion — the tree shape is load-bearing.
func TestDeleteGuardedByChildren(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}, hasChildren: true}
	uc := newDeptUC(rp)

	if _, err := uc.Delete(context.Background(), "d1"); !systempb.IsSystemDeptHasChildren(err) {
		t.Fatalf("has children: want DEPT_HAS_CHILDREN, got %v", err)
	}
	if len(rp.deletes) != 0 {
		t.Fatal("has children: guard must reject before the write")
	}
}

// A dept holding members refuses deletion.
func TestDeleteGuardedByUsers(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}, hasUsers: true}
	uc := newDeptUC(rp)

	if _, err := uc.Delete(context.Background(), "d1"); !systempb.IsSystemDeptHasUsers(err) {
		t.Fatalf("has users: want DEPT_HAS_USERS, got %v", err)
	}
	if len(rp.deletes) != 0 {
		t.Fatal("has users: guard must reject before the write")
	}
}

// An empty leaf deletes.
func TestDeleteHappy(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}}
	uc := newDeptUC(rp)

	if _, err := uc.Delete(context.Background(), "d1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deletes) != 1 {
		t.Fatal("delete: the row must be gone")
	}
}

// Members -------------------------------------------------------------------------------------------------------------

// Removing a member that is not in the dept answers one fuzzy not-found —
// absence covers moved-elsewhere, deleted, and out-of-scope alike.
func TestRemoveMemberFuzzyNotFound(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}, removed: false}
	uc := newDeptUC(rp)

	if err := uc.RemoveMember(context.Background(), "d1", "u1"); !systempb.IsSystemDeptMemberNotFound(err) {
		t.Fatalf("fuzzy: want DEPT_MEMBER_NOT_FOUND, got %v", err)
	}
}

// A present member is removed silently.
func TestRemoveMemberHappy(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}, removed: true}
	uc := newDeptUC(rp)

	if err := uc.RemoveMember(context.Background(), "d1", "u1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
}

// Listing members of an unknown dept rejects before the query fires.
func TestListMembersUnknownDept(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{}}
	uc := newDeptUC(rp)

	if _, err := uc.ListMembers(context.Background(), &MemberListIn{DeptID: "ghost"}); !systempb.IsSystemDeptNotFound(err) {
		t.Fatalf("unknown dept: want DEPT_NOT_FOUND, got %v", err)
	}
	if rp.membersCalls != 0 {
		t.Fatal("unknown dept: the query must not fire")
	}
}

// A legal reparent (no cycle, parent exists) goes through.
func TestUpdateReparentHappy(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}, "d2": {ID: "d2"}}, isAncestor: false}
	uc := newDeptUC(rp)

	if _, err := uc.Update(context.Background(), []string{"parent_id"}, &Dept{ID: "d1", ParentID: strPtr("d2")}); err != nil {
		t.Fatalf("reparent: %v", err)
	}
	if rp.ancestorCalled != 1 || rp.updates != 1 {
		t.Fatalf("reparent: ancestorCalled=%d updates=%d", rp.ancestorCalled, rp.updates)
	}
}

// List and Get are pure delegations to the port.
func TestListAndGetDelegate(t *testing.T) {
	rp := &fakeDeptRP{rows: map[string]*Dept{"d1": {ID: "d1"}}}
	uc := newDeptUC(rp)

	if _, err := uc.List(context.Background(), &DeptListIn{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := uc.Get(context.Background(), "d1"); err != nil {
		t.Fatalf("get: %v", err)
	}
}
