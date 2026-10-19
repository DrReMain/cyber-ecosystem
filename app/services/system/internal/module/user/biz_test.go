package user

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/module/population"
)

// passthroughTm runs the unit inline — these laws are about the UC's write
// order and side effects, not about transaction semantics.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeUserRP records the status/delete side: flips, updates, deletes, and the
// session kicks. A miss answers the classified DbNotFound error the real
// adapter would.
type fakeUserRP struct {
	UserRP
	rows  map[string]*User
	flips []struct {
		id      string
		enabled bool
	}
	updates   int
	updateErr error
	deletes   []string
	revokes   []string
	revokeErr error

	creates     int
	createdHash string
	// createErr is one-shot: the first Create fails (the account-unique race),
	// later calls succeed — the retry lands as a seat.
	createErr error

	// seats is the seat store; ensures records every EnsureSeat ask. An
	// existing seat is a no-op success, mirroring the adapter's DoNothing.
	seats         []*Seat
	ensures       []*Seat
	ensureErr     error
	deletedSeat   []string
	deletedCred   []string
	setRolesCalls []struct {
		userID  string
		roleIDs []string
	}
	setRolesErr    error
	notifyCalls    int
	listRolesCalls int
	roleIDsByCode  map[string]string
	rolesByUser    map[string][]string
	list           *UserListOut
	hadRoles       bool
	listRolesErr   error
	findAccountErr error
	notSeated      map[string]bool
}

// FindByAccount answers the registry-plane lookup: a row whose account matches,
// the classified not-found miss, or the injected infra failure.
func (f *fakeUserRP) FindByAccount(_ context.Context, account string) (*User, error) {
	if f.findAccountErr != nil {
		return nil, f.findAccountErr
	}
	for _, u := range f.rows {
		if u.Account != nil && *u.Account == account {
			return u, nil
		}
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) FindByID(_ context.Context, id string) (*User, error) {
	if u, ok := f.rows[id]; ok {
		return u, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) FindSeated(ctx context.Context, id string) (*User, error) {
	if f.notSeated[id] {
		return nil, errorspb.ErrorInfraErrorDbNotFound("")
	}
	return f.FindByID(ctx, id)
}

func (f *fakeUserRP) UpdateStatus(_ context.Context, id string, enabled bool) (*User, error) {
	f.flips = append(f.flips, struct {
		id      string
		enabled bool
	}{id, enabled})
	return f.rows[id], nil
}

func (f *fakeUserRP) Update(_ context.Context, _ []string, u *User) (*User, error) {
	f.updates++
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return u, nil
}

func (f *fakeUserRP) Delete(_ context.Context, id string) (string, error) {
	f.deletes = append(f.deletes, id)
	return id, nil
}

func (f *fakeUserRP) RevokeUserSessions(_ context.Context, userID string) error {
	f.revokes = append(f.revokes, userID)
	return f.revokeErr
}

func (f *fakeUserRP) NotifyChanged(context.Context) { f.notifyCalls++ }

func (f *fakeUserRP) DeleteUserRoles(_ context.Context, _ string) (bool, error) {
	return f.hadRoles, nil
}

func (f *fakeUserRP) Create(_ context.Context, u *User) (*User, error) {
	f.creates++
	if f.createErr != nil {
		err := f.createErr
		f.createErr = nil
		// The racing winner's row is what tripped the unique.
		if f.rows == nil {
			f.rows = map[string]*User{}
		}
		winner := *u
		winner.ID = "u1"
		f.rows[*u.Account] = &winner
		return nil, err
	}
	f.createdHash = utils.Deref(u.PasswordHash, "")
	out := *u
	out.ID = "new-u1"
	return &out, nil
}

func (f *fakeUserRP) EnsureSeat(_ context.Context, seat *Seat) error {
	f.ensures = append(f.ensures, seat)
	if f.ensureErr != nil {
		return f.ensureErr
	}
	for _, s := range f.seats {
		if *s == *seat {
			return nil
		}
	}
	f.seats = append(f.seats, seat)
	return nil
}

func (f *fakeUserRP) DeleteSeatsByUser(_ context.Context, userID string) error {
	f.deletedSeat = append(f.deletedSeat, userID)
	return nil
}

func (f *fakeUserRP) FindCredential(context.Context, string, string) (*Credential, error) {
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeUserRP) CreateCredential(context.Context, *Credential) error {
	return nil
}

func (f *fakeUserRP) DeleteCredentialsByUser(_ context.Context, userID string) error {
	f.deletedCred = append(f.deletedCred, userID)
	return nil
}

func (f *fakeUserRP) ListSeatsByUser(_ context.Context, _ string) ([]*Seat, error) {
	return f.seats, nil
}

func (f *fakeUserRP) SetUserRoles(_ context.Context, userID string, roleIDs []string) error {
	f.setRolesCalls = append(f.setRolesCalls, struct {
		userID  string
		roleIDs []string
	}{userID, roleIDs})
	return f.setRolesErr
}

func (f *fakeUserRP) ListUserRoles(_ context.Context, _ []string) (map[string][]string, error) {
	f.listRolesCalls++
	return f.rolesByUser, f.listRolesErr
}

func (f *fakeUserRP) FindRoleIDsByCodes(_ context.Context, codes []string) (map[string]string, error) {
	out := make(map[string]string, len(codes))
	for _, c := range codes {
		if id, ok := f.roleIDsByCode[c]; ok {
			out[c] = id
		}
	}
	return out, nil
}

func (f *fakeUserRP) List(_ context.Context, _ *UserListIn) (*UserListOut, error) {
	return f.list, nil
}

func newUserUCWith(rp *fakeUserRP, d dept.DeptRP, f file.FileRP, pop ...*fakePopRP) *UserUC {
	var pr population.PopulationRP = &fakePopRP{}
	if len(pop) > 0 {
		pr = pop[0]
	}
	return NewUserUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp, d, f, pr,
	)
}

func newUserUC(rp *fakeUserRP) *UserUC { return newUserUCWith(rp, nil, nil) }

// fakePopRP answers the registry lookup for birth-seat validation.
type fakePopRP struct {
	population.PopulationRP
	found map[string]*population.Population
}

func (f *fakePopRP) FindByCode(_ context.Context, code string) (*population.Population, error) {
	if p, ok := f.found[code]; ok {
		return p, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

// fakeDeptRP answers the dept reference check; fakeFileRP answers the avatar
// reference check with a controllable status.
type fakeDeptRP struct {
	dept.DeptRP
	err error
}

func (f *fakeDeptRP) Get(context.Context, string) (*dept.Dept, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &dept.Dept{}, nil
}

type fakeFileRP struct {
	file.FileRP
	ref *file.File
	err error
}

func (f *fakeFileRP) FindRef(context.Context, string) (*file.File, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ref, nil
}

func selfCtx(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID})
}

// UpdateStatus --------------------------------------------------------------------------------------------------------

// Disabling an account kicks its stored sessions (the disable itself must not
// be blocked by the kick's failure); enabling never touches sessions.
func TestUpdateStatusDisableRevokesSessions(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Enabled: true}}}
	uc := newUserUC(rp)

	if _, err := uc.UpdateStatus(context.Background(), "u1", false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if len(rp.revokes) != 1 || rp.revokes[0] != "u1" {
		t.Fatalf("disable: revokes = %v, want one kick for u1", rp.revokes)
	}

	// A failing kick is a warning: the disable itself stands.
	rp.revokes = nil
	rp.revokeErr = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, err := uc.UpdateStatus(context.Background(), "u1", true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if _, err := uc.UpdateStatus(context.Background(), "u1", false); err != nil {
		t.Fatalf("disable with dead kick: %v", err)
	}
	if len(rp.flips) != 3 {
		t.Fatalf("flips: want 3 status writes, got %d", len(rp.flips))
	}

	// Enabling never revokes: the enable flip above must have added no kick.
	if len(rp.revokes) != 1 {
		t.Fatalf("enable must not revoke: revokes = %v", rp.revokes)
	}
}

// Self-disable is the same lockout family as tenant self-disable: it rejects
// before any write side is reached.
func TestUpdateStatusSelfDisableGuarded(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Enabled: true}}}
	uc := newUserUC(rp)

	if _, err := uc.UpdateStatus(selfCtx("u1"), "u1", false); !systempb.IsSystemUserSelfDisable(err) {
		t.Fatalf("self-disable: want USER_SELF_DISABLE, got %v", err)
	}
	if len(rp.flips) != 0 || len(rp.revokes) != 0 {
		t.Fatal("self-disable: guard must reject before any write side")
	}
}

func TestUpdateStatusMissingClassified(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{}}
	uc := newUserUC(rp)

	if _, err := uc.UpdateStatus(context.Background(), "ghost", false); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("unknown id: want USER_NOT_FOUND, got %v", err)
	}
}

func TestByIdFacesRefuseForeignSeat(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Enabled: true}}, notSeated: map[string]bool{"u1": true}}
	uc := newUserUC(rp)

	if _, err := uc.Get(context.Background(), "u1"); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("get: want USER_NOT_FOUND, got %v", err)
	}
	if _, err := uc.Update(context.Background(), []string{"password"}, &User{ID: "u1"}, nil, nil); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("update: want USER_NOT_FOUND, got %v", err)
	}
	if _, err := uc.UpdateStatus(context.Background(), "u1", false); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("update status: want USER_NOT_FOUND, got %v", err)
	}
	if _, err := uc.Delete(context.Background(), "u1"); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("delete: want USER_NOT_FOUND, got %v", err)
	}
	if rp.updates != 0 || len(rp.flips) != 0 || len(rp.deletes) != 0 || len(rp.revokes) != 0 {
		t.Fatal("foreign seat: no write side may be reached")
	}
}

// Delete --------------------------------------------------------------------------------------------------------------

func TestDeleteSelfGuarded(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUC(rp)

	if _, err := uc.Delete(selfCtx("u1"), "u1"); !systempb.IsSystemUserSelfDelete(err) {
		t.Fatalf("self-delete: want USER_SELF_DELETE, got %v", err)
	}
	if len(rp.deletes) != 0 || len(rp.revokes) != 0 {
		t.Fatal("self-delete: guard must reject before any write side")
	}
}

func TestDeleteRevokesSessions(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUC(rp)

	if _, err := uc.Delete(context.Background(), "u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deletes) != 1 || len(rp.revokes) != 1 || rp.revokes[0] != "u1" {
		t.Fatalf("delete: deletes=%v revokes=%v, want both for u1", rp.deletes, rp.revokes)
	}
}

// ChangePassword ------------------------------------------------------------------------------------------------------

func TestChangePasswordOldMismatch(t *testing.T) {
	account := "a@x.co"
	oldHash, err := utils.Hash("right-old")
	if err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Account: &account, PasswordHash: &oldHash}}}
	uc := newUserUC(rp)

	if err := uc.ChangePassword(selfCtx("u1"), "wrong-old", "NewPass123"); !systempb.IsSystemPasswordMismatch(err) {
		t.Fatalf("old mismatch: want PASSWORD_MISMATCH, got %v", err)
	}
	if rp.updates != 0 {
		t.Fatal("old mismatch: no digest write may happen")
	}
}

// Create --------------------------------------------------------------------------------------------------------------

func TestCreateResolvesRefsAndRoles(t *testing.T) {
	rp := &fakeUserRP{roleIDsByCode: map[string]string{"admin": "r1"}}
	uc := newUserUCWith(rp,
		&fakeDeptRP{},
		&fakeFileRP{ref: &file.File{Status: file.StatusConfirmed}},
	)

	account, password, deptID, avatar := "a@x.co", "pw-123456", "d1", "f1"
	out, err := uc.Create(context.Background(), &account, &password, &deptID, &avatar, "", []string{"admin"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if out.ID != "new-u1" || rp.creates != 1 {
		t.Fatalf("create: out=%+v creates=%d", out, rp.creates)
	}
	if utils.NeedsRehash(rp.createdHash) {
		t.Fatal("create: digest must be argon2id")
	}
	if len(rp.setRolesCalls) != 1 || rp.setRolesCalls[0].roleIDs[0] != "r1" {
		t.Fatalf("create: setRoles = %+v", rp.setRolesCalls)
	}
	if rp.notifyCalls != 1 {
		t.Fatalf("create: notify = %d, want 1", rp.notifyCalls)
	}
}

func TestCreateRejectsUnknownDept(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, &fakeDeptRP{err: errorspb.ErrorInfraErrorDbNotFound("")}, nil)

	account, password, deptID := "a@x.co", "pw-123456", "d1"
	if _, err := uc.Create(context.Background(), &account, &password, &deptID, nil, "", nil); !systempb.IsSystemDeptNotFound(err) {
		t.Fatalf("unknown dept: want DEPT_NOT_FOUND, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unknown dept: no row may be created")
	}
}

func TestCreateRejectsUnknownAvatar(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, nil, &fakeFileRP{err: errorspb.ErrorInfraErrorDbNotFound("")})

	account, password, avatar := "a@x.co", "pw-123456", "ghost"
	if _, err := uc.Create(context.Background(), &account, &password, nil, &avatar, "", nil); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("unknown avatar: want FILE_NOT_FOUND, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unknown avatar: no row may be created")
	}
}

func TestCreateRejectsUnconfirmedAvatar(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, nil, &fakeFileRP{ref: &file.File{Status: "uploading"}})

	account, password, avatar := "a@x.co", "pw-123456", "f1"
	if _, err := uc.Create(context.Background(), &account, &password, nil, &avatar, "", nil); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("unconfirmed avatar: want FILE_INVALID_STATE, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unconfirmed avatar: no row may be created")
	}
}

func TestCreateRejectsUnknownRole(t *testing.T) {
	rp := &fakeUserRP{roleIDsByCode: map[string]string{}}
	uc := newUserUC(rp)

	account, password := "a@x.co", "pw-123456"
	if _, err := uc.Create(context.Background(), &account, &password, nil, nil, "", []string{"ghost"}); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("unknown role: want ROLE_NOT_FOUND, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("unknown role: no row may be created")
	}
}

func TestCreateSetRolesFailureSurfaces(t *testing.T) {
	rp := &fakeUserRP{roleIDsByCode: map[string]string{"admin": "r1"}, setRolesErr: errorspb.ErrorInfraErrorDbConstraint("")}
	uc := newUserUC(rp)

	account, password := "a@x.co", "pw-123456"
	if _, err := uc.Create(context.Background(), &account, &password, nil, nil, "", []string{"admin"}); !errorspb.IsInfraErrorDbConstraint(err) {
		t.Fatalf("setRoles failure: want infra DbConstraint passthrough, got %v", err)
	}
	if rp.notifyCalls != 0 {
		t.Fatal("setRoles failure: NotifyChanged must not fire")
	}
}

// Update --------------------------------------------------------------------------------------------------------------

func TestUpdateFullMask(t *testing.T) {
	rp := &fakeUserRP{roleIDsByCode: map[string]string{"admin": "r1"}, rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUCWith(rp,
		&fakeDeptRP{},
		&fakeFileRP{ref: &file.File{Status: file.StatusConfirmed}},
	)

	password := "new-pw-123"
	out, err := uc.Update(context.Background(),
		[]string{"dept_id", "avatar", "password", "roles"},
		&User{ID: "u1"}, &password, []string{"admin"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if out == nil || out.ID != "u1" || rp.updates != 1 {
		t.Fatalf("update: out=%v updates=%d", out, rp.updates)
	}
	if len(rp.setRolesCalls) != 1 || rp.notifyCalls != 1 {
		t.Fatalf("update: setRoles=%v notify=%d", rp.setRolesCalls, rp.notifyCalls)
	}
}

func TestUpdatePasswordOnly(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUC(rp)

	if _, err := uc.Update(context.Background(), []string{"password"}, &User{ID: "u1"}, nil, nil); err != nil {
		t.Fatalf("update: %v", err)
	}
	if rp.updates != 1 || len(rp.setRolesCalls) != 0 || rp.notifyCalls != 0 {
		t.Fatalf("password-only: updates=%d setRoles=%v notify=%d", rp.updates, rp.setRolesCalls, rp.notifyCalls)
	}
}

// A constraint tripped by an update is an infra passthrough — the account is
// immutable and births classify their own races at the Create choke.
func TestUpdateConstraintStaysInfra(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUC(rp)
	rp.updateErr = errorspb.ErrorInfraErrorDbConstraint("")

	if _, err := uc.Update(context.Background(), []string{"dept_id"}, &User{ID: "u1"}, nil, nil); !errorspb.IsInfraErrorDbConstraint(err) {
		t.Fatalf("dept mask: want infra passthrough, got %v", err)
	}
	if len(rp.setRolesCalls) != 0 || rp.notifyCalls != 0 {
		t.Fatalf("failed updates must not touch bindings: setRoles=%v notify=%d", rp.setRolesCalls, rp.notifyCalls)
	}
}

// List / Get / BatchGet ------------------------------------------------------------------------------------------------

func TestListHydratesRolesAndEmptyExitsEarly(t *testing.T) {
	rp := &fakeUserRP{
		list:        &UserListOut{List: []*User{{ID: "u1"}, {ID: "u2"}}},
		rolesByUser: map[string][]string{"u1": {"admin"}, "u2": nil},
	}
	uc := newUserUC(rp)

	out, err := uc.List(context.Background(), &UserListIn{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if out.List[0].Roles[0] != "admin" || rp.listRolesCalls != 1 {
		t.Fatalf("list: roles=%v calls=%d", out.List[0].Roles, rp.listRolesCalls)
	}

	rp.list = &UserListOut{List: nil}
	rp.listRolesCalls = 0
	if _, err := uc.List(context.Background(), &UserListIn{}); err != nil {
		t.Fatalf("empty list: %v", err)
	}
	if rp.listRolesCalls != 0 {
		t.Fatal("empty list: the role query must not fire")
	}
}

func TestGetHydratesRoles(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}, rolesByUser: map[string][]string{"u1": {"admin"}}}
	uc := newUserUC(rp)

	out, err := uc.Get(context.Background(), "u1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(out.Roles) != 1 || out.Roles[0] != "admin" {
		t.Fatalf("get: roles = %v", out.Roles)
	}
}

func TestBatchGetDelegates(t *testing.T) {
	rp := &fakeUserRP{UserRP: &stubFindByIDs{}}
	uc := newUserUC(rp)

	if _, err := uc.BatchGet(context.Background(), []string{"u1"}); err != nil {
		t.Fatalf("batch: %v", err)
	}
}

type stubFindByIDs struct {
	UserRP
	called bool
}

func (s *stubFindByIDs) FindByIDs(context.Context, []string) (map[string]*User, error) {
	s.called = true
	return nil, nil
}

// ChangePassword ------------------------------------------------------------------------------------------------------

func TestChangePasswordRewritesDigest(t *testing.T) {
	account := "a@x.co"
	oldHash, err := utils.Hash("right-old")
	if err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Account: &account, PasswordHash: &oldHash}}, updates: 0}
	uc := newUserUC(rp)

	if err := uc.ChangePassword(selfCtx("u1"), "right-old", "NewPass123"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if rp.updates != 1 {
		t.Fatalf("change password: updates = %d, want 1", rp.updates)
	}
}

func TestUpdateRejectsUnknownDept(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUCWith(rp, &fakeDeptRP{err: errorspb.ErrorInfraErrorDbNotFound("")}, nil)

	deptID := "ghost"
	if _, err := uc.Update(context.Background(), []string{"dept_id"}, &User{ID: "u1", DeptID: &deptID}, nil, nil); !systempb.IsSystemDeptNotFound(err) {
		t.Fatalf("update unknown dept: want DEPT_NOT_FOUND, got %v", err)
	}
	if rp.updates != 0 {
		t.Fatal("update unknown dept: no write may happen")
	}
}

func TestChangePasswordUnknownSelf(t *testing.T) {
	uc := newUserUC(&fakeUserRP{rows: map[string]*User{}})

	if err := uc.ChangePassword(selfCtx("ghost"), "old", "NewPass123"); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("unknown self: want USER_NOT_FOUND, got %v", err)
	}
}

func TestDeleteNotifiesWhenRolesExist(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}, hadRoles: true}
	uc := newUserUC(rp)

	if _, err := uc.Delete(context.Background(), "u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if rp.notifyCalls != 1 {
		t.Fatalf("delete: notify = %d, want 1 when bindings existed", rp.notifyCalls)
	}
}

func TestGetUnknownClassified(t *testing.T) {
	uc := newUserUC(&fakeUserRP{rows: map[string]*User{}})

	if _, err := uc.Get(context.Background(), "ghost"); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("get unknown: want USER_NOT_FOUND, got %v", err)
	}
}

func TestListRolesErrorSurfaces(t *testing.T) {
	rp := &fakeUserRP{
		list:         &UserListOut{List: []*User{{ID: "u1"}}},
		listRolesErr: errorspb.ErrorInfraErrorNetworkConnection(""),
	}
	uc := newUserUC(rp)

	if _, err := uc.List(context.Background(), &UserListIn{}); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("roles error: want infra passthrough, got %v", err)
	}
}

// Membership birth/death ---------------------------------------------------------------------------------------------

func TestCreateBirthsSeat(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, nil, nil)
	account, pw := "a@x.co", "pw123456"

	out, err := uc.Create(context.Background(), &account, &pw, nil, nil, "", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(rp.seats) != 1 {
		t.Fatalf("birth: seats = %v, want exactly one", rp.seats)
	}
	seat := rp.seats[0]
	if seat.UserID != out.ID || seat.Population != security.DefaultPopulation || seat.TenantID != security.DefaultTenant {
		t.Fatalf("birth: seat = %+v (user %s)", seat, out.ID)
	}
}

func TestCreatePopulationSeatBorn(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, nil, nil, &fakePopRP{found: map[string]*population.Population{
		"guest": {Code: "guest", Plane: security.PlaneMember},
	}})
	account, pw := "b@x.co", "pw123456"

	if _, err := uc.Create(context.Background(), &account, &pw, nil, nil, "guest", nil); err != nil {
		t.Fatalf("create: %v", err)
	}
	seat := rp.seats[0]
	if seat.Population != "guest" {
		t.Fatalf("birth: seat population = %q, want guest", seat.Population)
	}
}

func TestCreateRejectsUnknownPopulation(t *testing.T) {
	rp := &fakeUserRP{}
	uc := newUserUCWith(rp, nil, nil)
	account, pw := "c@x.co", "pw123456"

	if _, err := uc.Create(context.Background(), &account, &pw, nil, nil, "ghost", nil); !systempb.IsSystemPopulationNotFound(err) {
		t.Fatalf("unknown population: want POPULATION_NOT_FOUND, got %v", err)
	}
	if rp.creates != 0 || len(rp.seats) != 0 {
		t.Fatal("unknown population: no row, no seat")
	}
}

// Seat ----------------------------------------------------------------------------------------------------------------

func TestCreateSeatsExistingIdentity(t *testing.T) {
	twin := "twin@x.co"
	rp := &fakeUserRP{
		rows:          map[string]*User{"u1": {ID: "u1", Account: &twin}},
		roleIDsByCode: map[string]string{"admin": "r1"},
	}
	uc := newUserUCWith(rp, nil, nil)
	account, password := twin, "pw123456"
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "admin1", TenantID: "acme"})

	out, err := uc.Create(ctx, &account, &password, nil, nil, "", []string{"admin"})
	if err != nil {
		t.Fatalf("seat: %v", err)
	}
	if out.ID != "u1" {
		t.Fatalf("seat: out = %+v, want the existing identity", out)
	}
	if rp.creates != 0 {
		t.Fatal("seat: no second row may be born")
	}
	if len(rp.ensures) != 1 || rp.ensures[0].UserID != "u1" || rp.ensures[0].TenantID != "acme" ||
		rp.ensures[0].Population != security.DefaultPopulation {
		t.Fatalf("seat: ensured = %+v, want {u1 acme %s}", rp.ensures, security.DefaultPopulation)
	}
	if len(rp.setRolesCalls) != 1 || rp.setRolesCalls[0].userID != "u1" {
		t.Fatalf("seat: roles = %v, want bound for u1 in the seat's tenant", rp.setRolesCalls)
	}
}

func TestCreateSeatWithoutRoleAskLeavesBindings(t *testing.T) {
	twin := "twin@x.co"
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Account: &twin}}}
	uc := newUserUCWith(rp, nil, nil)
	account, password := twin, "pw123456"
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "admin1", TenantID: "acme"})

	if _, err := uc.Create(ctx, &account, &password, nil, nil, "", nil); err != nil {
		t.Fatalf("seat: %v", err)
	}
	if len(rp.ensures) != 1 {
		t.Fatalf("seat: ensured = %+v, want one seat", rp.ensures)
	}
	if len(rp.setRolesCalls) != 0 {
		t.Fatalf("seat: roles = %v, want no binding touch", rp.setRolesCalls)
	}
}

func TestCreateSeatCarriesPopulation(t *testing.T) {
	twin := "twin@x.co"
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1", Account: &twin}}}
	uc := newUserUCWith(rp, nil, nil, &fakePopRP{found: map[string]*population.Population{
		"guest": {Code: "guest", Plane: security.PlaneMember},
	}})
	account, password := twin, "pw123456"
	ctx := security.WithSubject(context.Background(), &security.Subject{TenantID: "default"})

	if _, err := uc.Create(ctx, &account, &password, nil, nil, "guest", nil); err != nil {
		t.Fatalf("seat population: %v", err)
	}
	if len(rp.seats) != 1 || rp.seats[0].Population != "guest" {
		t.Fatalf("seat population: seats = %+v, want guest", rp.seats)
	}
}

func TestCreateDuplicateSeatCosmetic(t *testing.T) {
	twin := "twin@x.co"
	rp := &fakeUserRP{
		rows:          map[string]*User{"u1": {ID: "u1", Account: &twin}},
		roleIDsByCode: map[string]string{"admin": "r1"},
		seats: []*Seat{
			{UserID: "u1", TenantID: security.DefaultTenant, Population: security.DefaultPopulation},
		},
	}
	uc := newUserUCWith(rp, nil, nil)
	account, password := twin, "pw123456"

	out, err := uc.Create(context.Background(), &account, &password, nil, nil, "", []string{"admin"})
	if err != nil {
		t.Fatalf("cosmetic: %v", err)
	}
	if out.ID != "u1" {
		t.Fatalf("cosmetic: out = %+v", out)
	}
	if len(rp.seats) != 1 {
		t.Fatalf("cosmetic: store = %v, want the one existing seat", rp.seats)
	}
	if len(rp.setRolesCalls) != 1 {
		t.Fatal("cosmetic: the role binding still runs")
	}
}

func TestCreateBirthRaceLosesToSeat(t *testing.T) {
	twin := "race@x.co"
	rp := &fakeUserRP{
		rows:      map[string]*User{},
		createErr: errorspb.ErrorInfraErrorDbConstraint(""),
	}
	uc := newUserUCWith(rp, nil, nil)
	account, password := twin, "pw123456"

	out, err := uc.Create(context.Background(), &account, &password, nil, nil, "", nil)
	if err != nil {
		t.Fatalf("race loser: %v", err)
	}
	if rp.creates != 1 || out.ID != "u1" {
		t.Fatalf("race loser: creates=%d out=%+v, want one failed birth then the winner's seat", rp.creates, out)
	}
	if len(rp.seats) != 1 || rp.seats[0].UserID != "u1" || rp.seats[0].TenantID != security.DefaultTenant {
		t.Fatalf("race loser: seats = %+v, want the winner seated in default", rp.seats)
	}
}

func TestCreateAccountLookupFailureSurfaces(t *testing.T) {
	rp := &fakeUserRP{findAccountErr: errorspb.ErrorInfraErrorDbInternal("")}
	uc := newUserUCWith(rp, nil, nil)
	account, password := "twin@x.co", "pw123456"

	if _, err := uc.Create(context.Background(), &account, &password, nil, nil, "", nil); !errorspb.IsInfraErrorDbInternal(err) {
		t.Fatalf("lookup failure: err = %v, want passthrough", err)
	}
	if rp.creates != 0 || len(rp.seats) != 0 {
		t.Fatal("lookup failure: no row, no seat")
	}
}

func TestDeleteKillsSeats(t *testing.T) {
	rp := &fakeUserRP{rows: map[string]*User{"u1": {ID: "u1"}}}
	uc := newUserUCWith(rp, nil, nil)

	if _, err := uc.Delete(context.Background(), "u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deletedSeat) != 1 || rp.deletedSeat[0] != "u1" {
		t.Fatalf("death: deleted = %v, want [u1]", rp.deletedSeat)
	}
}
