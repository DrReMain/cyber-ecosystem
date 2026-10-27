package e2e

// The tenant write-path guard, esports twin of the system one: the read
// interceptor narrows queries and the insert hook back-fills tenant_id,
// while updates and deletes must carry a subject-tenant predicate at the
// mutation layer itself. ActivationCode is hard-deleted, so a raw delete
// keeps its native delete op — the form a soft-delete rewrite would hide —
// and both mutation forms narrow under one predicate. The precedence
// mirrors the read interceptor (explicit domain switch > subject >
// no-subject skip), and the two exemptions are pinned as intended so
// neither the enforcement nor its escape hatches drift silently. Gated:
// ENT_E2E=1 with the dev database up runs it; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	entactivationcode "cyber-ecosystem/app/services/esports/internal/ent/activationcode"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

const twTenB = "e2e-tw-tenant-b"

func twEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func twSubject(tenant string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: "e2etwsubject", TenantID: tenant, Population: "user"})
}

func cleanTenantWrite(t *testing.T, db *ent.Client) {
	t.Helper()
	if _, err := db.ActivationCode.Delete().Where(entactivationcode.Code("e2e-tw-code")).Exec(context.Background()); err != nil {
		t.Fatalf("cleanup activation_code: %v", err)
	}
}

func TestTenantWriteGuardLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the tenant write-path end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: twEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: twEnvOr("DB_PASS", "postgres"), DbName: twEnvOr("DB_NAME", "esports"),
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
	ctxA := twSubject("e2e-tw-tenant-a")

	// ActivationCode carries the tenant mixin and no soft-delete mixin; the
	// insert hook takes the tenant from the creating subject, so the row
	// starts pinned to tenant B. used_count is the mutable guard field.
	codeB, err := db.ActivationCode.Create().SetCode("e2e-tw-code").SetReportType("cognitive").
		SetExpiresAt(time.Now().Add(24 * time.Hour)).SetCreatedBy("e2e-tw-admin").Save(twSubject(twTenB))
	if err != nil {
		t.Fatalf("fixture activation_code: %v", err)
	}

	mustCount := func(want int) {
		t.Helper()
		got, rerr := db.ActivationCode.Query().Where(entactivationcode.IDEQ(codeB.ID)).Only(bare)
		if rerr != nil {
			t.Fatalf("reload activation_code: %v", rerr)
		}
		if got.UsedCount != want || got.TenantID != twTenB {
			t.Fatalf("activation_code drifted: used=%d tenant=%q want used=%d tenant %q",
				got.UsedCount, got.TenantID, want, twTenB)
		}
	}

	// Arm 1 — cross-tenant raw update: subject A points a bare UpdateOneID at
	// tenant B's row; the predicate must leave it untouched (zero rows).
	_, err = db.ActivationCode.UpdateOneID(codeB.ID).SetUsedCount(3).Save(ctxA)
	if !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant update must answer NotFound, got: %v", err)
	}
	mustCount(0)

	// Arm 2 — cross-tenant raw delete: no soft-delete rewrite here, the
	// native delete op itself must carry the predicate.
	if err = db.ActivationCode.DeleteOneID(codeB.ID).Exec(ctxA); !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant delete must answer NotFound, got: %v", err)
	}
	mustCount(0)

	// Arm 3 — the explicit domain-switch exemption: WithTenant outranks the
	// subject, so a tenant-B switch on a tenant-A subject may write tenant
	// B's row.
	ctxSwitch := security.WithTenant(ctxA, twTenB)
	if _, err = db.ActivationCode.UpdateOneID(codeB.ID).SetUsedCount(1).Save(ctxSwitch); err != nil {
		t.Fatalf("explicit-switch update must pass, got: %v", err)
	}
	mustCount(1)

	// Arm 4 — the subject-less exemption: contexts carrying no subject get
	// no predicate and the write goes through.
	if _, err = db.ActivationCode.UpdateOneID(codeB.ID).SetUsedCount(2).Save(bare); err != nil {
		t.Fatalf("subject-less update must pass, got: %v", err)
	}
	mustCount(2)
}
