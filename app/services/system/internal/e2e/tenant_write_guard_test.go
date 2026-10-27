package e2e

// The tenant write-path guard. The read interceptor narrows queries and the
// insert hook back-fills tenant_id, but until the mutation-layer predicate
// existed, updates and deletes rode the biz-layer pre-read convention alone —
// a correctly-written module could not be confused by a raw cross-tenant
// UpdateOneID, yet nothing structural answered for it. This guard pins the
// structural law at the mutation layer itself: a subject's tenant predicate
// rides every update and delete with the same precedence as the read
// interceptor (explicit domain switch > subject > no-subject skip), and the
// two exemptions — the explicit domain switch and the subject-less sweep —
// are pinned as intended, so neither the enforcement nor its escape hatches
// drift silently. Row-level asserts carry the proof: a refused write must
// leave the foreign row byte-identical, whatever error shape the zero-row
// update surfaces. Gated: ENT_E2E=1 with the dev database up runs it;
// anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entrole "cyber-ecosystem/app/services/system/internal/ent/role"
	local_mixins "cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	twTenA = "e2e-tw-tenant-a"
	twTenB = "e2e-tw-tenant-b"
)

func twEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func twSubject(tenant string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "e2etwsubject", TenantID: tenant, Population: "user"})
}

// cleanTenantWrite removes every fixture row by prefix; the lift is moot on
// the hard-delete role table but keeps the sweep shape portable to the
// soft-deleted tenant-mixin tables a future reader-guard may pin here.
func cleanTenantWrite(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.Role.Delete().Where(entrole.CodeHasPrefix("e2e-tw-")).Exec(ctx); err != nil {
		t.Fatalf("cleanup role: %v", err)
	}
}

func TestTenantWriteGuardLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the tenant write-path end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: twEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: twEnvOr("DB_PASS", "postgres"), DbName: twEnvOr("DB_NAME", "system"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	cleanTenantWrite(t, db)
	t.Cleanup(func() {
		cleanTenantWrite(t, db)
		closeDB()
	})

	bare := context.Background()
	ctxA := twSubject(twTenA)

	// The role table carries the tenant mixin but no soft-delete mixin — the
	// exact shape where a raw delete stays a delete (no update rewrite), so
	// it pins both mutation forms. The insert hook takes the tenant from the
	// creating subject, so the row starts pinned to tenant B.
	roleB, err := db.Role.Create().SetCode("e2e-tw-role-b").SetName("e2e role B").Save(twSubject(twTenB))
	if err != nil {
		t.Fatalf("fixture role B: %v", err)
	}

	mustName := func(want string) {
		t.Helper()
		got, rerr := db.Role.Query().Where(entrole.IDEQ(roleB.ID)).Only(bare)
		if rerr != nil {
			t.Fatalf("reload role B: %v", rerr)
		}
		if got.Name != want || got.TenantID != twTenB {
			t.Fatalf("role B drifted: name=%q tenant=%q want name=%q tenant %q",
				got.Name, got.TenantID, want, twTenB)
		}
	}

	// Arm 1 — cross-tenant raw update: subject A points a bare UpdateOneID at
	// tenant B's row; the predicate must leave it untouched (zero rows).
	_, err = db.Role.UpdateOneID(roleB.ID).SetName("e2e-tw-hijacked").Save(ctxA)
	if !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant update must answer NotFound, got: %v", err)
	}
	mustName("e2e role B")

	// Arm 2 — cross-tenant raw delete: the same subject tries to delete the
	// foreign row. Tenant-mixin tables come in both flavors — soft-deleted
	// ones rewrite the delete into an update that re-enters the hook chain,
	// hard ones keep the native delete op — and the predicate must narrow
	// both; role here exercises the native form.
	if err = db.Role.DeleteOneID(roleB.ID).Exec(ctxA); !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant delete must answer NotFound, got: %v", err)
	}
	mustName("e2e role B")

	// Arm 3 — the explicit domain-switch exemption: WithTenant outranks the
	// subject (one precedence axis for reads, buckets, and writes), so a
	// tenant-B switch on a tenant-A subject may write tenant B's row.
	ctxSwitch := security.WithTenant(ctxA, twTenB)
	if _, err = db.Role.UpdateOneID(roleB.ID).SetName("e2e-tw-switch-pass").Save(ctxSwitch); err != nil {
		t.Fatalf("explicit-switch update must pass, got: %v", err)
	}
	mustName("e2e-tw-switch-pass")

	// Arm 4 — the subject-less exemption: registry/sweep contexts carry no
	// subject, so no predicate applies and the write goes through.
	if _, err = db.Role.UpdateOneID(roleB.ID).SetName("e2e-tw-bare-pass").Save(bare); err != nil {
		t.Fatalf("subject-less update must pass, got: %v", err)
	}
	mustName("e2e-tw-bare-pass")
}
