package e2e

// The authz stamp's three laws as executable counterexamples. Truth is
// the authz_stamps row, bumped inside each change transaction by the
// local bump mixin. Law one: every mutation of a snapshot-feeding table
// moves the stamp, and user rows move it only when they cross the
// disabled set — never on profile-only edits, which would storm every
// replica. Law two: a replica that missed the wake still converges — one
// publish and the reconcile pulls the snapshot to the stamp. Law three
// is the regression this slice fixes: a stamp that moved BACKWARDS
// rebuilds instead of pinning the old snapshot, which is what the old >=
// comparison did forever after a counter loss or a restore. Gated:
// ENT_E2E=1 with the dev stack (postgres + redis) up runs it.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entpermission "cyber-ecosystem/app/services/system/internal/ent/permission"
	entpopulation "cyber-ecosystem/app/services/system/internal/ent/population"
	entrole "cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	enttenant "cyber-ecosystem/app/services/system/internal/ent/tenant"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	stampTenant  = "e2e-authz-tenant"
	stampRole2   = "e2e-authz-role2"
	stampProbeOp = "e2e.authz/probe/*"
	stampProbe2  = "e2e.authz/probe2/*"
)

func stampRead(t *testing.T, db *ent.Client) int64 {
	t.Helper()
	row, err := db.AuthzStamp.Get(context.Background(), local_mixins.StampRowKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0
		}
		t.Fatalf("stamp read: %v", err)
	}
	return row.Version
}

// cleanAuthzStampE2E rides the soft-delete lift so user tombstones from
// crashed runs sweep too; the cleanup's own hook bumps are harmless —
// the stamp only needs to move, and assertions are relative.
func cleanAuthzStampE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.Permission.Delete().Where(entpermission.OperationHasPrefix("e2e.authz/")).Exec(ctx); err != nil {
		t.Fatalf("cleanup permissions: %v", err)
	}
	if _, err := db.Population.Delete().Where(entpopulation.CodeHasPrefix("e2e-authz-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup populations: %v", err)
	}
	if _, err := db.Role.Delete().Where(entrole.CodeHasPrefix("e2e-authz-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup roles: %v", err)
	}
	if _, err := db.User.Delete().Where(entuser.AccountHasPrefix("e2e-authz-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup users: %v", err)
	}
	if _, err := db.Tenant.Delete().Where(enttenant.CodeEQ(stampTenant)).Exec(ctx); err != nil {
		t.Fatalf("cleanup tenant: %v", err)
	}
}

func stampWaitFor(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", d)
}

func TestAuthzStampLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the authz-stamp end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	bare := context.Background()

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: tenEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: tenEnvOr("DB_PASS", "postgres"), DbName: tenEnvOr("DB_NAME", "system"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	cacheClient, closeCache, err := platform.NewCache(&conf.Data{Redis: &conf.Data_Redis{
		Addr: tenEnvOr("REDIS_ADDR", "localhost:6379"), Db: 0,
	}}, logger)
	if err != nil {
		t.Fatalf("cache: %v", err)
	}
	pf, err := platform.NewPlatform(cacheClient, nil, db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanAuthzStampE2E(t, db)
	t.Cleanup(func() {
		cleanAuthzStampE2E(t, db)
		closeCache()
		closeDB()
	})

	if _, err = db.Tenant.Create().SetCode(stampTenant).SetName("e2e authz stamp").Save(bare); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	ctxT := tenSubject(stampTenant)

	// Law one — every mutation of a snapshot-feeding table moves the stamp.
	s0 := stampRead(t, db)
	role1, err := db.Role.Create().SetCode("e2e-authz-role1").SetName("e2e role one").Save(ctxT)
	if err != nil {
		t.Fatalf("role create: %v", err)
	}
	if s1 := stampRead(t, db); s1 <= s0 {
		t.Fatalf("role create must bump: %d → %d", s0, s1)
	}
	if _, err = db.Role.UpdateOneID(role1.ID).SetName("renamed").Save(ctxT); err != nil {
		t.Fatalf("role update: %v", err)
	}
	if s2 := stampRead(t, db); s2 <= s0 {
		t.Fatalf("role update must bump: %d → %d", s0, s2)
	}
	if _, err = db.Role.Delete().Where(entrole.IDEQ(role1.ID)).Exec(ctxT); err != nil {
		t.Fatalf("role delete: %v", err)
	}
	if s3 := stampRead(t, db); s3 <= s0 {
		t.Fatalf("role delete must bump: %d → %d", s0, s3)
	}

	// Law one, the narrow user gate — profile edits never bump; the
	// disabled-set crossing and the disappearance do.
	u, err := db.User.Create().SetAccount("e2e-authz-user1").SetPasswordHash("e2e-not-a-digest").Save(bare)
	if err != nil {
		t.Fatalf("user create: %v", err)
	}
	afterUser := stampRead(t, db)
	if _, err = db.User.UpdateOneID(u.ID).SetPasswordHash("e2e-another-not-a-digest").Save(bare); err != nil {
		t.Fatalf("user profile edit: %v", err)
	}
	if s := stampRead(t, db); s != afterUser {
		t.Fatalf("profile-only edit must not bump: %d → %d", afterUser, s)
	}
	if _, err = db.User.UpdateOneID(u.ID).SetEnabled(false).Save(bare); err != nil {
		t.Fatalf("user disable: %v", err)
	}
	if s := stampRead(t, db); s <= afterUser {
		t.Fatalf("disable must bump: %d → %d", afterUser, s)
	}
	if err = db.User.DeleteOneID(u.ID).Exec(bare); err != nil {
		t.Fatalf("user delete: %v", err)
	}
	if s := stampRead(t, db); s <= afterUser {
		t.Fatalf("user delete must bump: %d → %d", afterUser, s)
	}

	// Law one, the tombstone lift — a restore clears the marker through
	// the SET-NULL path and the user re-enters the snapshot, so the
	// un-delete moves the stamp exactly as the tombstone did.
	u2, err := db.User.Create().SetAccount("e2e-authz-user2").SetPasswordHash("e2e-not-a-digest").Save(bare)
	if err != nil {
		t.Fatalf("user create two: %v", err)
	}
	afterCreate := stampRead(t, db)
	if err = db.User.DeleteOneID(u2.ID).Exec(bare); err != nil {
		t.Fatalf("user soft delete: %v", err)
	}
	if s := stampRead(t, db); s <= afterCreate {
		t.Fatalf("soft delete must bump: %d → %d", afterCreate, s)
	}
	if _, err = db.User.UpdateOneID(u2.ID).ClearDeletedAt().Save(local_mixins.SkipSoftDelete(context.Background())); err != nil {
		t.Fatalf("user restore: %v", err)
	}
	if s := stampRead(t, db); s <= afterCreate {
		t.Fatalf("restore must bump: %d → %d", afterCreate, s)
	}

	// Law one, the vocabulary table — a population change moves the stamp
	// because the snapshot compiles code → plane from it.
	beforePop := stampRead(t, db)
	if _, err = db.Population.Create().SetCode("e2e-authz-pop").SetName("e2e pop").SetPlane("backend").Save(bare); err != nil {
		t.Fatalf("population create: %v", err)
	}
	if s := stampRead(t, db); s <= beforePop {
		t.Fatalf("population create must bump: %d → %d", beforePop, s)
	}

	// Laws two and three need a live replica with a real pubsub watch.
	if _, err = db.Role.Create().SetCode(stampRole2).SetName("e2e role two").Save(ctxT); err != nil {
		t.Fatalf("role2 create: %v", err)
	}
	if _, err = db.Permission.Create().SetRoleID(lookupRoleID(t, db, stampRole2)).SetOperation(stampProbeOp).Save(ctxT); err != nil {
		t.Fatalf("permission create: %v", err)
	}
	rp, err := authzmod.NewAuthzRP(logger, pf)
	if err != nil {
		t.Fatalf("authz rp: %v", err)
	}
	if err = rp.StartWatch(bare); err != nil {
		t.Fatalf("start watch: %v", err)
	}
	t.Cleanup(func() { _ = rp.StopWatch(bare) })

	preview := func() []string {
		ops, perr := rp.PreviewGrants(bare, stampTenant, []string{stampRole2})
		if perr != nil {
			t.Fatalf("preview: %v", perr)
		}
		return ops
	}
	has := func(list []string, op string) bool {
		for _, s := range list {
			if s == op {
				return true
			}
		}
		return false
	}
	if !has(preview(), stampProbeOp) {
		t.Fatalf("initial compile must see the probe grant")
	}

	// Law two — a missed wake stays stale until the publish, then heals.
	if _, err = db.Permission.Delete().Where(entpermission.OperationEQ(stampProbeOp)).Exec(ctxT); err != nil {
		t.Fatalf("permission delete: %v", err)
	}
	if !has(preview(), stampProbeOp) {
		t.Fatalf("stale replica expected before the wake")
	}
	shared.NotifyPolicyChanged(bare, cacheClient, logger)
	stampWaitFor(t, 3*time.Second, func() bool { return !has(preview(), stampProbeOp) })

	// Law three — the regression: the stamp moves BACKWARDS while the
	// replica holds a higher watermark. The old >= comparison skipped
	// forever; != rebuilds, and the change that never woke the replica
	// becomes visible.
	if _, err = db.Permission.Create().SetRoleID(lookupRoleID(t, db, stampRole2)).SetOperation(stampProbe2).Save(ctxT); err != nil {
		t.Fatalf("permission2 create: %v", err)
	}
	if has(preview(), stampProbe2) {
		t.Fatalf("stale replica expected: the wake was missed")
	}
	if err = db.AuthzStamp.UpdateOneID(local_mixins.StampRowKey).SetVersion(1).Exec(bare); err != nil {
		t.Fatalf("stamp rewind: %v", err)
	}
	shared.NotifyPolicyChanged(bare, cacheClient, logger)
	stampWaitFor(t, 3*time.Second, func() bool { return has(preview(), stampProbe2) })
}

func lookupRoleID(t *testing.T, db *ent.Client, code string) string {
	t.Helper()
	r, err := db.Role.Query().Where(entrole.CodeEQ(code)).Only(context.Background())
	if err != nil {
		t.Fatalf("lookup role %s: %v", code, err)
	}
	return r.ID
}
