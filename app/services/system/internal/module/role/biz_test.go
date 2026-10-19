package role

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/module/policy"
	"cyber-ecosystem/app/services/system/internal/module/resource"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// fakeResourceRP serves the catalog the pattern validation consults — a
// single service with one method is all the vocabulary these laws need.
type fakeResourceRP struct {
	resource.ResourceRP
	catalog []*resource.ServiceMeta
}

func (f *fakeResourceRP) ListResource(context.Context) ([]*resource.ServiceMeta, error) {
	return f.catalog, nil
}

func newRoleUC(catalog []*resource.ServiceMeta) *RoleUC {
	return &RoleUC{
		UC:         shared.NewUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil),
		resourceRP: &fakeResourceRP{catalog: catalog},
	}
}

func seedCatalog() []*resource.ServiceMeta {
	return []*resource.ServiceMeta{{
		Name:     "Svc",
		FullName: "cyber.example.Svc",
		Methods:  []*resource.ResourceMethod{{Name: "Method"}},
	}}
}

// validateOperation is save-time honesty: a pattern is legal only if the
// catalog can back it — global bypasses by definition, wildcards need the
// service, exact needs the method; a non-tail "*" is malformed, never a
// wildcard.
func TestValidateOperationPatterns(t *testing.T) {
	uc := newRoleUC(seedCatalog())
	cases := []struct {
		pattern string
		wantErr bool
	}{
		{"/*", false},                        // global bypasses the catalog by definition
		{"/cyber.example.Svc/*", false},      // wildcard: service must exist
		{"/cyber.example.Svc/Method", false}, // exact: method must exist
		{"/cyber.example.Svc/Ghost", true},   // exact: unknown method
		{"/cyber.ghost.Svc/*", true},         // wildcard: unknown service
		{"no-leading-slash", true},           // malformed
		{"/cyber.example.Svc/*x", true},      // a "*" that is not a clean tail is not a wildcard
		{"/nodot/*", true},                   // a service segment carries a package dot
	}
	for _, c := range cases {
		err := uc.validateOperation(context.Background(), c.pattern)
		if got := err != nil; got != c.wantErr {
			t.Fatalf("validate(%q) err=%v, wantErr=%v", c.pattern, err, c.wantErr)
		}
		if err != nil && !systempb.IsSystemPermissionInvalidOperation(err) {
			t.Fatalf("validate(%q): want PERMISSION_INVALID_OPERATION, got %v", c.pattern, err)
		}
	}
}

// passthroughTm runs the unit inline — these laws are about write order and
// broadcast conditions, not transaction semantics.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeRoleRP records the write side: creates, grant replacements, deletes,
// broadcasts; rows back the FindByID checks.
type fakeRoleRP struct {
	RoleRP
	rows         map[string]*Role
	creates      int
	replaced     []string
	deleted      []string
	notifyCalls  int
	hasBindings  bool
	grantsByRole map[string][]*Permission
	boundCount   int64
	membersOut   *MemberListOut
	membersCalls int
	findErr      error
}

func (f *fakeRoleRP) FindByID(_ context.Context, id string) (*Role, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeRoleRP) Create(_ context.Context, r *Role) (*Role, error) {
	f.creates++
	out := *r
	out.ID = "r-new"
	return &out, nil
}

func (f *fakeRoleRP) Update(_ context.Context, _ []string, r *Role) (*Role, error) {
	return r, nil
}

func (f *fakeRoleRP) UpdateStatus(_ context.Context, id string, _ bool) (*Role, error) {
	return f.rows[id], nil
}

func (f *fakeRoleRP) Delete(_ context.Context, id string) (string, error) {
	f.deleted = append(f.deleted, id)
	return id, nil
}

func (f *fakeRoleRP) ExistsBindings(context.Context, string) (bool, error) {
	return f.hasBindings, nil
}

func (f *fakeRoleRP) ReplaceGrants(_ context.Context, roleID string, _ []*Permission) error {
	f.replaced = append(f.replaced, roleID)
	return nil
}

func (f *fakeRoleRP) ListGrants(_ context.Context, roleID string) ([]*Permission, error) {
	return f.grantsByRole[roleID], nil
}

func (f *fakeRoleRP) CountBindings(context.Context, string) (int64, error) {
	return f.boundCount, nil
}

func (f *fakeRoleRP) NotifyChanged(context.Context) { f.notifyCalls++ }

func (f *fakeRoleRP) RemoveMember(context.Context, string, string, string) error { return nil }

func (f *fakeRoleRP) ListMembers(_ context.Context, _ *MemberListIn) (*MemberListOut, error) {
	f.membersCalls++
	return f.membersOut, nil
}

// fakePolicyRP answers the policy reference check with a controllable found
// set.
type fakePolicyRP struct {
	policy.PolicyRP
	found []*policy.Policy
	err   error
}

func (f *fakePolicyRP) FindByIDs(context.Context, []string) ([]*policy.Policy, error) {
	return f.found, f.err
}

// fakeAuthzRP captures the tenant the diagnosis was scoped to.
type fakeAuthzRP struct {
	authz.AuthzRP
	explain      *authz.ExplainResult
	preview      []string
	err          error
	captureScope []string
}

func (f *fakeAuthzRP) ExplainGrants(_ context.Context, _, _, tenant, _ string) (*authz.ExplainResult, error) {
	f.captureScope = append(f.captureScope, tenant)
	return f.explain, f.err
}

func (f *fakeAuthzRP) PreviewGrants(_ context.Context, tenant string, _ []string) ([]string, error) {
	f.captureScope = append(f.captureScope, tenant)
	return f.preview, f.err
}

func newRoleUCFull(rp *fakeRoleRP, prp *fakePolicyRP, arp *fakeAuthzRP, cat []*resource.ServiceMeta) *RoleUC {
	return &RoleUC{
		UC:         shared.NewUC(slog.New(slog.NewTextHandler(io.Discard, nil)), passthroughTm{}),
		roleRP:     rp,
		resourceRP: &fakeResourceRP{catalog: cat},
		authzRP:    arp,
		policyRP:   prp,
	}
}

func ctxWithTenant(userID, tenant string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID, TenantID: tenant})
}

// Create --------------------------------------------------------------------------------------------------------------

// A valid grant set lands role + grants in one transaction and fans out one
// rebuild broadcast.
func TestCreateReplacesGrantsAndNotifies(t *testing.T) {
	rp := &fakeRoleRP{}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Create(context.Background(), &Role{Code: "ops"}, []*Permission{{Operation: "/cyber.example.Svc/Method"}}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if rp.creates != 1 || len(rp.replaced) != 1 || rp.notifyCalls != 1 {
		t.Fatalf("create: creates=%d replaced=%v notify=%d", rp.creates, rp.replaced, rp.notifyCalls)
	}
}

// A duplicate operation inside one grant set is rejected before any write.
func TestCreateRejectsDuplicateOperations(t *testing.T) {
	rp := &fakeRoleRP{}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	dup := []*Permission{
		{Operation: "/cyber.example.Svc/Method"},
		{Operation: "/cyber.example.Svc/Method"},
	}
	if _, err := uc.Create(context.Background(), &Role{Code: "ops"}, dup); !systempb.IsSystemPermissionInvalidOperation(err) {
		t.Fatalf("duplicate: want PERMISSION_INVALID_OPERATION, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("duplicate: no write may happen")
	}
}

// A policy id that resolves to nothing is rejected with the missing ids in
// the cause — absence covers deleted rows and other-tenant rows alike.
func TestCreateRejectsUnknownPolicy(t *testing.T) {
	rp := &fakeRoleRP{}
	uc := newRoleUCFull(rp, &fakePolicyRP{found: nil}, &fakeAuthzRP{}, seedCatalog())

	g := []*Permission{{Operation: "/cyber.example.Svc/Method", PolicyIDs: []string{"p-ghost"}}}
	if _, err := uc.Create(context.Background(), &Role{Code: "ops"}, g); !systempb.IsSystemPolicyNotFound(err) {
		t.Fatalf("unknown policy: want POLICY_NOT_FOUND, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unknown policy: no write may happen")
	}
}

// A grant set whose policy ids all resolve passes the reference check.
func TestCreateWithPoliciesHappy(t *testing.T) {
	rp := &fakeRoleRP{}
	uc := newRoleUCFull(rp, &fakePolicyRP{found: []*policy.Policy{{ID: "p1"}}}, &fakeAuthzRP{}, seedCatalog())

	g := []*Permission{{Operation: "/cyber.example.Svc/Method", PolicyIDs: []string{"p1"}}}
	if _, err := uc.Create(context.Background(), &Role{Code: "ops"}, g); err != nil {
		t.Fatalf("create with policies: %v", err)
	}
	if rp.creates != 1 {
		t.Fatal("create with policies: the row must be born")
	}
}

// Update --------------------------------------------------------------------------------------------------------------

// An unknown role rejects before any write side.
func TestUpdateUnknownRole(t *testing.T) {
	uc := newRoleUCFull(&fakeRoleRP{rows: map[string]*Role{}}, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Update(context.Background(), []string{"name"}, &Role{ID: "ghost"}, nil); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("unknown role: want ROLE_NOT_FOUND, got %v", err)
	}
}

// Only grants feed the compiled snapshot: a cosmetic edit skips the rebuild
// broadcast entirely.
func TestUpdateCosmeticSkipsNotify(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Update(context.Background(), []string{"name"}, &Role{ID: "r1", Name: "n"}, nil); err != nil {
		t.Fatalf("cosmetic update: %v", err)
	}
	if len(rp.replaced) != 0 || rp.notifyCalls != 0 {
		t.Fatalf("cosmetic update: replaced=%v notify=%d — both must be zero", rp.replaced, rp.notifyCalls)
	}
}

// A grants mask rebinds in the same transaction and broadcasts the rebuild.
func TestUpdateGrantsMaskNotifies(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	g := []*Permission{{Operation: "/cyber.example.Svc/Method"}}
	if _, err := uc.Update(context.Background(), []string{"grants"}, &Role{ID: "r1"}, g); err != nil {
		t.Fatalf("grants update: %v", err)
	}
	if len(rp.replaced) != 1 || rp.notifyCalls != 1 {
		t.Fatalf("grants update: replaced=%v notify=%d", rp.replaced, rp.notifyCalls)
	}
}

// UpdateStatus / Delete ------------------------------------------------------------------------------------------------

// A status flip always broadcasts — a disabled role loses its policy rows on
// the next compile.
func TestUpdateStatusNotifies(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.UpdateStatus(context.Background(), "r1", false); err != nil {
		t.Fatalf("status flip: %v", err)
	}
	if rp.notifyCalls != 1 {
		t.Fatalf("status flip: notify = %d, want 1", rp.notifyCalls)
	}
}

// A role with live bindings refuses deletion — the engine snapshot depends
// on the principal_role rows.
func TestDeleteGuardedByBindings(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}, hasBindings: true}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Delete(context.Background(), "r1"); !systempb.IsSystemRoleHasBindings(err) {
		t.Fatalf("bound delete: want ROLE_HAS_BINDINGS, got %v", err)
	}
	if len(rp.deleted) != 0 || rp.notifyCalls != 0 {
		t.Fatal("bound delete: guard must reject before any write")
	}
}

// An unbound role deletes and broadcasts.
func TestDeleteHappy(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Delete(context.Background(), "r1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deleted) != 1 || rp.notifyCalls != 1 {
		t.Fatalf("delete: deleted=%v notify=%d", rp.deleted, rp.notifyCalls)
	}
}

// Get ------------------------------------------------------------------------------------------------------------------

// An unknown role classifies as ROLE_NOT_FOUND — a domain shape, not the
// infra identity; a generic store error keeps its infra identity.
func TestGetUnknownClassifiedGenericPassthrough(t *testing.T) {
	uc := newRoleUCFull(&fakeRoleRP{rows: map[string]*Role{}}, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.Get(context.Background(), "ghost"); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("unknown role: want ROLE_NOT_FOUND, got %v", err)
	}

	uc2 := newRoleUCFull(&fakeRoleRP{findErr: errorspb.ErrorInfraErrorNetworkConnection("")}, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())
	if _, err := uc2.Get(context.Background(), "r1"); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("generic error: want infra passthrough, got %v", err)
	}
}

// Detail hydrates grants with the same light counts the list shows — global,
// wildcard, narrowing, and policy tallies agree between faces.
func TestGetHydratesCountsAndBindings(t *testing.T) {
	rp := &fakeRoleRP{
		rows: map[string]*Role{"r1": {ID: "r1"}},
		grantsByRole: map[string][]*Permission{
			"r1": {
				{Operation: "/*", ScopeKind: ""},
				{Operation: "/cyber.example.Svc/*", ScopeKind: kauthz.ScopeKindSelf, PolicyIDs: []string{"p1"}},
			},
		},
		boundCount: 7,
	}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	out, err := uc.Get(context.Background(), "r1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.GrantCount != 2 || !out.GlobalGrant || out.WildcardCount != 1 || out.NarrowedCount != 1 || out.PolicyCount != 1 || out.BoundCount != 7 {
		t.Fatalf("get counts: %+v", out)
	}
}

// Members / explain / preview ------------------------------------------------------------------------------------------

// Listing members of an unknown role rejects before the query fires.
func TestListMembersUnknownRole(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if _, err := uc.ListMembers(context.Background(), &MemberListIn{RoleID: "ghost"}); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("members unknown role: want ROLE_NOT_FOUND, got %v", err)
	}
	if rp.membersCalls != 0 {
		t.Fatal("members unknown role: the query must not fire")
	}
}

// Removing a member broadcasts the rebuild — bindings feed the snapshot too.
func TestRemoveMemberNotifies(t *testing.T) {
	rp := &fakeRoleRP{rows: map[string]*Role{"r1": {ID: "r1"}}}
	uc := newRoleUCFull(rp, &fakePolicyRP{}, &fakeAuthzRP{}, seedCatalog())

	if err := uc.RemoveMember(context.Background(), "r1", "user", "u1"); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	if rp.notifyCalls != 1 {
		t.Fatalf("remove member: notify = %d, want 1", rp.notifyCalls)
	}
}

// The diagnosis faces are caller-scoped: no subject is unauthenticated, the
// engine is called with the caller's own tenant, and engine failure is
// INTERNAL — never a diagnosis outcome.
func TestExplainScopedToCallerTenant(t *testing.T) {
	arp := &fakeAuthzRP{explain: &authz.ExplainResult{Allowed: true}}
	uc := newRoleUCFull(&fakeRoleRP{}, &fakePolicyRP{}, arp, seedCatalog())

	if _, err := uc.ExplainOperation(context.Background(), "user", "u1", "/op"); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no subject: want unauthenticated, got %v", err)
	}

	if _, err := uc.ExplainOperation(ctxWithTenant("me", "t9"), "user", "u1", "/op"); err != nil {
		t.Fatalf("explain: %v", err)
	}
	if len(arp.captureScope) != 1 || arp.captureScope[0] != "t9" {
		t.Fatalf("explain: scope = %v, want the caller's tenant", arp.captureScope)
	}

	arpErr := &fakeAuthzRP{err: errorspb.ErrorInfraErrorNetworkConnection("")}
	uc2 := newRoleUCFull(&fakeRoleRP{}, &fakePolicyRP{}, arpErr, seedCatalog())
	if _, err := uc2.ExplainOperation(ctxWithTenant("me", "t9"), "user", "u1", "/op"); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("engine failure: want INTERNAL, got %v", err)
	}
}

// Preview mirrors explain: subject-bound, tenant-scoped, engine failure is
// infra.
func TestPreviewScopedToCallerTenant(t *testing.T) {
	arp := &fakeAuthzRP{preview: []string{"/svc/*"}}
	uc := newRoleUCFull(&fakeRoleRP{}, &fakePolicyRP{}, arp, seedCatalog())

	if _, err := uc.PreviewGrants(context.Background(), nil); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no subject: want unauthenticated, got %v", err)
	}

	ops, err := uc.PreviewGrants(ctxWithTenant("me", "t9"), []string{"ops"})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if len(ops) != 1 || arp.captureScope[0] != "t9" {
		t.Fatalf("preview: ops=%v scope=%v", ops, arp.captureScope)
	}
}
