package e2e

// The tenant write-path guard, agent twin of the system one: the read
// interceptor narrows queries and the insert hook back-fills tenant_id,
// while updates and deletes must carry a subject-tenant predicate at the
// mutation layer itself — agent_config is soft-deleted, so a raw delete is
// rewritten into an update that re-enters the hook chain, and both mutation
// forms narrow under one predicate. The precedence mirrors the read
// interceptor (explicit domain switch > subject > no-subject skip), and the
// two exemptions are pinned as intended so neither the enforcement nor its
// escape hatches drift silently. Gated: ENT_E2E=1 with the dev database up
// runs it; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/ent"
	entagentconfig "cyber-ecosystem/app/services/agent/internal/ent/agentconfig"
	local_mixins "cyber-ecosystem/app/services/agent/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/agent/internal/platform"
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

// cleanTenantWrite removes fixture rows by prefix, riding the soft-delete
// lift so tombstones from crashed runs are swept and the live-only unique
// index on user_id frees up.
func cleanTenantWrite(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.AgentConfig.Delete().Where(entagentconfig.UserID("e2e-tw-user")).Exec(ctx); err != nil {
		t.Fatalf("cleanup agent_config: %v", err)
	}
}

func TestTenantWriteGuardLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the tenant write-path end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: twEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: twEnvOr("DB_PASS", "postgres"), DbName: twEnvOr("DB_NAME", "agent"),
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

	// AgentConfig carries the tenant mixin; the insert hook takes the tenant
	// from the creating subject, so the row starts pinned to tenant B.
	cfgB, err := db.AgentConfig.Create().SetUserID("e2e-tw-user").SetBaseURL("https://e2e.example").Save(twSubject(twTenB))
	if err != nil {
		t.Fatalf("fixture agent_config: %v", err)
	}

	mustURL := func(want string) {
		t.Helper()
		got, rerr := db.AgentConfig.Query().Where(entagentconfig.IDEQ(cfgB.ID)).Only(bare)
		if rerr != nil {
			t.Fatalf("reload agent_config: %v", rerr)
		}
		if got.BaseURL != want || got.TenantID != twTenB || got.DeletedAt != nil {
			t.Fatalf("agent_config drifted: url=%q tenant=%q deleted=%v want url=%q live %q",
				got.BaseURL, got.TenantID, got.DeletedAt, want, twTenB)
		}
	}

	// Arm 1 — cross-tenant raw update: subject A points a bare UpdateOneID at
	// tenant B's row; the predicate must leave it untouched (zero rows).
	_, err = db.AgentConfig.UpdateOneID(cfgB.ID).SetBaseURL("https://e2e-tw-hijacked").Save(ctxA)
	if !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant update must answer NotFound, got: %v", err)
	}
	mustURL("https://e2e.example")

	// Arm 2 — cross-tenant raw delete: the soft-delete rewrite turns the
	// delete into an update that re-enters the hook chain, and the same
	// predicate must narrow it; the row must stay live.
	if err = db.AgentConfig.DeleteOneID(cfgB.ID).Exec(ctxA); !ent.IsNotFound(err) {
		t.Fatalf("cross-tenant delete must answer NotFound, got: %v", err)
	}
	mustURL("https://e2e.example")

	// Arm 3 — the explicit domain-switch exemption: WithTenant outranks the
	// subject, so a tenant-B switch on a tenant-A subject may write tenant
	// B's row.
	ctxSwitch := security.WithTenant(ctxA, twTenB)
	if _, err = db.AgentConfig.UpdateOneID(cfgB.ID).SetBaseURL("https://e2e-tw-switch.example").Save(ctxSwitch); err != nil {
		t.Fatalf("explicit-switch update must pass, got: %v", err)
	}
	mustURL("https://e2e-tw-switch.example")

	// Arm 4 — the subject-less exemption: contexts carrying no subject get
	// no predicate and the write goes through.
	if _, err = db.AgentConfig.UpdateOneID(cfgB.ID).SetBaseURL("https://e2e-tw-bare.example").Save(bare); err != nil {
		t.Fatalf("subject-less update must pass, got: %v", err)
	}
	mustURL("https://e2e-tw-bare.example")
}
