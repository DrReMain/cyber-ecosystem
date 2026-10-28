package e2e

// The tenant-isolation live guard, in the cross-module live-e2e home: this
// directory holds nothing but env-gated end-to-end tests (no non-test
// files, so `go build` skips the package and it can never leak into a
// production import). Tenancy on the read path is enforced by three
// mechanisms, and this guard pins each one where it lives: role and file
// rows carry the tenant mixin (query interceptor filters, insert hook
// back-fills); user rows carry no tenant column — visibility narrows to
// the caller's seats (membership join); audit rows carry a plain tenant
// field — the read rule is stamped by the UC onto the list input. The
// write path is enforced in two layers: the UC's tenant-filtered pre-read
// answers NotFound before any mutation (the refusal asserted below —
// preserving not-found semantics and anti-enumeration), and the mixin's
// update/delete predicate backstops it at the mutation layer (pinned by
// the write-guard test in this home). The subject-less context is
// asserted too, as the intended exemption (registry sweeps), so nobody
// tightens or "fixes" it by accident.
// Gated: ENT_E2E=1 with the dev database up runs it; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entauditlog "cyber-ecosystem/app/services/system/internal/ent/auditlog"
	entfile "cyber-ecosystem/app/services/system/internal/ent/file"
	entmembership "cyber-ecosystem/app/services/system/internal/ent/membership"
	entrole "cyber-ecosystem/app/services/system/internal/ent/role"
	local_mixins "cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	enttenant "cyber-ecosystem/app/services/system/internal/ent/tenant"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	auditmod "cyber-ecosystem/app/services/system/internal/module/audit"
	filemod "cyber-ecosystem/app/services/system/internal/module/file"
	rolemod "cyber-ecosystem/app/services/system/internal/module/role"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	tenA = "e2e-tenant-a"
	tenB = "e2e-tenant-b"
)

func tenEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func tenSubject(tenant string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "e2etensubject", TenantID: tenant, Population: "user"})
}

func tenPage() *commonpb.PageRequest {
	pageNo, size := int32(1), int32(100)
	return &commonpb.PageRequest{PageNo: &pageNo, PageSize: &size}
}

// cleanTenantE2E removes every fixture row by prefix, children before tenants.
// It rides the soft-delete lift so tombstones from crashed runs are seen and
// hard-deleted too — the sweep must leave nothing behind.
func cleanTenantE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.AuditLog.Delete().Where(entauditlog.OperationHasPrefix("e2e ")).Exec(ctx); err != nil {
		t.Fatalf("cleanup audit: %v", err)
	}
	if _, err := db.File.Delete().Where(entfile.KeyHasPrefix("e2e/")).Exec(ctx); err != nil {
		t.Fatalf("cleanup file: %v", err)
	}
	if _, err := db.Role.Delete().Where(entrole.CodeHasPrefix("e2e-tenant-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup role: %v", err)
	}
	if _, err := db.Membership.Delete().Where(entmembership.TenantIDIn(tenA, tenB)).Exec(ctx); err != nil {
		t.Fatalf("cleanup membership: %v", err)
	}
	if _, err := db.User.Delete().Where(entuser.AccountHasPrefix("e2e-tenant-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup user: %v", err)
	}
	if _, err := db.Tenant.Delete().Where(enttenant.CodeHasPrefix("e2e-tenant-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup tenant: %v", err)
	}
}

// tenStampAuditRP captures the tenant the UC stamps onto the list input; the
// remaining faces panic — they are not under test here.
type tenStampAuditRP struct{ captured string }

// tenHooks absorbs the FileUC constructor's lifecycle registrations; the
// sweep loops never start in this guard.
type tenHooks struct{}

func (tenHooks) OnStart(func(context.Context) error) {}
func (tenHooks) OnStop(func(context.Context) error)  {}

func (s *tenStampAuditRP) ListAuditLogs(_ context.Context, in *auditmod.AuditListIn) (*auditmod.AuditListOut, error) {
	s.captured = in.Tenant
	return &auditmod.AuditListOut{}, nil
}
func (*tenStampAuditRP) Publish(context.Context, *audit.Event) { panic("not under test") }
func (*tenStampAuditRP) CreateExport(context.Context, string, string) error {
	panic("not under test")
}
func (*tenStampAuditRP) ListExports(context.Context, *auditmod.ExportListIn) (*auditmod.ExportListOut, error) {
	panic("not under test")
}
func (*tenStampAuditRP) ListLogsAfter(context.Context, string, string, int) ([]*auditmod.AuditLog, error) {
	panic("not under test")
}
func (*tenStampAuditRP) FindLogsByIDs(context.Context, string, []string) ([]*auditmod.AuditLog, error) {
	panic("not under test")
}

func TestTenantIsolationLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the tenant-isolation end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: tenEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: tenEnvOr("DB_PASS", "postgres"), DbName: tenEnvOr("DB_NAME", "system"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	pf, err := platform.NewPlatform(nil, nil, db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanTenantE2E(t, db)
	t.Cleanup(func() {
		cleanTenantE2E(t, db)
		closeDB()
	})

	bare := context.Background()
	ctxA, ctxB := tenSubject(tenA), tenSubject(tenB)

	// Fixtures. Roles and files take their tenant from the subject (the
	// insert hook); users are global identities seated per tenant; audit
	// rows state their tenant directly.
	if _, err = db.Tenant.Create().SetCode(tenA).SetName("e2e tenant A").Save(bare); err != nil {
		t.Fatalf("tenant A: %v", err)
	}
	if _, err = db.Tenant.Create().SetCode(tenB).SetName("e2e tenant B").Save(bare); err != nil {
		t.Fatalf("tenant B: %v", err)
	}
	roleA, err := db.Role.Create().SetCode(tenA + "-role").SetName("e2e role A").Save(ctxA)
	if err != nil {
		t.Fatalf("role A: %v", err)
	}
	roleB, err := db.Role.Create().SetCode(tenB + "-role").SetName("e2e role B").Save(ctxB)
	if err != nil {
		t.Fatalf("role B: %v", err)
	}
	fileA, err := db.File.Create().SetKey("e2e/" + tenA + "/f1").SetName("f1").SetStatus("confirmed").Save(ctxA)
	if err != nil {
		t.Fatalf("file A: %v", err)
	}
	fileB, err := db.File.Create().SetKey("e2e/" + tenB + "/f1").SetName("f1").SetStatus("confirmed").Save(ctxB)
	if err != nil {
		t.Fatalf("file B: %v", err)
	}
	userA, err := db.User.Create().SetAccount("e2e-tenant-a-user1").SetPasswordHash("e2ehash-not-a-real-digest").Save(bare)
	if err != nil {
		t.Fatalf("user A: %v", err)
	}
	userB, err := db.User.Create().SetAccount("e2e-tenant-b-user1").SetPasswordHash("e2ehash-not-a-real-digest").Save(bare)
	if err != nil {
		t.Fatalf("user B: %v", err)
	}
	if _, err = db.Membership.Create().SetUserID(userA.ID).SetTenantID(tenA).SetPopulation("user").Save(bare); err != nil {
		t.Fatalf("membership A: %v", err)
	}
	if _, err = db.Membership.Create().SetUserID(userB.ID).SetTenantID(tenB).SetPopulation("user").Save(bare); err != nil {
		t.Fatalf("membership B: %v", err)
	}
	if _, err = db.AuditLog.Create().SetTenant(tenA).SetOperation("e2e op A").SetEventID("e2e-event-a").Save(bare); err != nil {
		t.Fatalf("audit A: %v", err)
	}
	if _, err = db.AuditLog.Create().SetTenant(tenB).SetOperation("e2e op B").SetEventID("e2e-event-b").Save(bare); err != nil {
		t.Fatalf("audit B: %v", err)
	}

	// Positive controls: the owning tenant reads its own rows — proving the
	// refusals below come from isolation, not broken fixtures.
	roleUC := rolemod.NewRoleUC(logger, pf, rolemod.NewRoleRP(logger, pf), nil, nil, nil)
	fileUC := filemod.NewFileUC(logger, pf, filemod.NewFileRP(logger, pf), nil, tenHooks{})
	userUC := usermod.NewUserUC(logger, pf, usermod.NewUserRP(logger, pf), nil, nil, nil)
	if _, err = roleUC.Get(ctxB, roleB.ID); err != nil {
		t.Fatalf("positive control role: %v", err)
	}
	if _, err = fileUC.Get(ctxB, fileB.ID); err != nil {
		t.Fatalf("positive control file: %v", err)
	}
	if _, err = userUC.Get(ctxB, userB.ID); err != nil {
		t.Fatalf("positive control user: %v", err)
	}

	// Arm 1 — list narrowing: tenant A's faces never see tenant B's rows.
	roles, err := roleUC.List(ctxA, &rolemod.RoleListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("role list: %v", err)
	}
	roleCodes := map[string]bool{}
	for _, r := range roles.List {
		roleCodes[r.Code] = true
	}
	if !roleCodes[roleA.Code] || roleCodes[roleB.Code] {
		t.Fatalf("role list narrowing: %v", roleCodes)
	}
	files, err := fileUC.List(ctxA, &filemod.FileListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("file list: %v", err)
	}
	fileKeys := map[string]bool{}
	for _, f := range files.List {
		fileKeys[f.Key] = true
	}
	if !fileKeys[fileA.Key] || fileKeys[fileB.Key] {
		t.Fatalf("file list narrowing: %v", fileKeys)
	}
	users, err := userUC.List(ctxA, &usermod.UserListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("user list: %v", err)
	}
	userIDs := map[string]bool{}
	for _, u := range users.List {
		userIDs[u.ID] = true
	}
	if !userIDs[userA.ID] || userIDs[userB.ID] {
		t.Fatalf("user list narrowing: %v", userIDs)
	}

	// Arm 2 — by-ID refusal: a foreign row by its real ID answers NotFound.
	if _, err = roleUC.Get(ctxA, roleB.ID); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("role by-ID refusal: %v", err)
	}
	if _, err = fileUC.Get(ctxA, fileB.ID); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("file by-ID refusal: %v", err)
	}
	if _, err = userUC.Get(ctxA, userB.ID); !systempb.IsSystemUserNotFound(err) {
		t.Fatalf("user by-ID refusal: %v", err)
	}

	// Arm 3 — write refusal: the biz convention (tenant-filtered pre-read)
	// answers NotFound before any mutation, and the foreign row survives.
	if _, err = roleUC.UpdateStatus(ctxA, roleB.ID, false); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("role write refusal (status): %v", err)
	}
	if _, err = roleUC.Delete(ctxA, roleB.ID); !systempb.IsSystemRoleNotFound(err) {
		t.Fatalf("role write refusal (delete): %v", err)
	}
	if _, err = fileUC.Update(ctxA, fileB.ID, "renamed"); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("file write refusal (update): %v", err)
	}
	if err = fileUC.Delete(ctxA, fileB.ID); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("file write refusal (delete): %v", err)
	}
	if n, cerr := db.Role.Query().Where(entrole.IDEQ(roleB.ID)).Count(bare); cerr != nil || n != 1 {
		t.Fatalf("foreign role survived refused writes: n=%d err=%v", n, cerr)
	}
	if n, cerr := db.File.Query().Where(entfile.IDEQ(fileB.ID)).Count(bare); cerr != nil || n != 1 {
		t.Fatalf("foreign file survived refused writes: n=%d err=%v", n, cerr)
	}

	// Arm 4 — the subject-less exemption: registry/sweep contexts see rows
	// of both tenants (contains, not equals — the dev registry has more).
	rolesAll, err := roleUC.List(bare, &rolemod.RoleListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("bare role list: %v", err)
	}
	roleCodesAll := map[string]bool{}
	for _, r := range rolesAll.List {
		roleCodesAll[r.Code] = true
	}
	if !roleCodesAll[roleA.Code] || !roleCodesAll[roleB.Code] {
		t.Fatalf("subject-less exemption (role): %v", roleCodesAll)
	}
	filesAll, err := fileUC.List(bare, &filemod.FileListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("bare file list: %v", err)
	}
	fileKeysAll := map[string]bool{}
	for _, f := range filesAll.List {
		fileKeysAll[f.Key] = true
	}
	if !fileKeysAll[fileA.Key] || !fileKeysAll[fileB.Key] {
		t.Fatalf("subject-less exemption (file): %v", fileKeysAll)
	}
	usersAll, err := userUC.List(bare, &usermod.UserListIn{PageRequest: tenPage()})
	if err != nil {
		t.Fatalf("bare user list: %v", err)
	}
	userIDsAll := map[string]bool{}
	for _, u := range usersAll.List {
		userIDsAll[u.ID] = true
	}
	if !userIDsAll[userA.ID] || !userIDsAll[userB.ID] {
		t.Fatalf("subject-less exemption (user): %v", userIDsAll)
	}

	// Audit — the read rule is the UC stamping the caller's tenant onto the
	// list input (the data-side tenant filter consumes exactly this field).
	stub := &tenStampAuditRP{}
	auditUC := auditmod.NewAuditUC(logger, pf, stub, nil)
	if _, err = auditUC.List(ctxA, &auditmod.AuditListIn{PageRequest: tenPage()}); err != nil {
		t.Fatalf("audit list A: %v", err)
	}
	if stub.captured != tenA {
		t.Fatalf("audit stamp A: %q", stub.captured)
	}
	if _, err = auditUC.List(ctxB, &auditmod.AuditListIn{PageRequest: tenPage()}); err != nil {
		t.Fatalf("audit list B: %v", err)
	}
	if stub.captured != tenB {
		t.Fatalf("audit stamp B: %q", stub.captured)
	}
}
