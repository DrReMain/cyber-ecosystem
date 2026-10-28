package e2e

// The session-liveness live guard: verification freshness rides the engine
// stamp. The mint embeds the version it saw; an unchanged version answers
// from the single-row comparison alone, and a disable still lands on the
// very next request because the bump rides the mutating transaction
// itself. The gates are checks, not revocations — a re-enable resurrects
// the stored key, now re-stamped by the write-back. Every hop runs on the
// live stack: real mixin hooks, real stamp row, real session store.
// Gated: ENT_E2E=1 with the dev stack (postgres + redis) up runs it.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	entmembership "cyber-ecosystem/app/services/system/internal/ent/membership"
	entpopulation "cyber-ecosystem/app/services/system/internal/ent/population"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	enttenant "cyber-ecosystem/app/services/system/internal/ent/tenant"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
	authmod "cyber-ecosystem/app/services/system/internal/module/auth"
	authzmod "cyber-ecosystem/app/services/system/internal/module/authz"
	populationmod "cyber-ecosystem/app/services/system/internal/module/population"
	tenantmod "cyber-ecosystem/app/services/system/internal/module/tenant"
	usermod "cyber-ecosystem/app/services/system/internal/module/user"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	liveTenant  = "e2e-liveness-tenant"
	liveAccount = "e2e-liveness@e2e"
)

// cleanLivenessE2E rides the soft-delete lift so user tombstones from
// crashed runs sweep too; the cleanup's own bumps are harmless — the stamp
// only needs to move, and assertions are relative.
func cleanLivenessE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := local_mixins.SkipSoftDelete(context.Background())
	if _, err := db.Membership.Delete().Where(entmembership.TenantIDEQ(liveTenant)).Exec(ctx); err != nil {
		t.Fatalf("cleanup membership: %v", err)
	}
	if _, err := db.User.Delete().Where(entuser.AccountHasPrefix("e2e-liveness")).Exec(ctx); err != nil {
		t.Fatalf("cleanup user: %v", err)
	}
	if _, err := db.Tenant.Delete().Where(enttenant.CodeEQ(liveTenant)).Exec(ctx); err != nil {
		t.Fatalf("cleanup tenant: %v", err)
	}
}

func TestSessionLivenessLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the session-liveness end-to-end")
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
	cacheErrHandler, err := platform.NewCacheErrorHandler()
	if err != nil {
		t.Fatalf("cache error handler: %v", err)
	}
	pf, err := platform.NewPlatform(cacheClient, cacheErrHandler, db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanLivenessE2E(t, db)
	t.Cleanup(func() {
		cleanLivenessE2E(t, db)
		closeCache()
		closeDB()
	})

	// The default population's row feeds the engine's plane registry, so it
	// must predate the snapshot compile; a leftover from a prior run (or a
	// seeded database) is fine as-is.
	popSeen, err := db.Population.Query().Where(entpopulation.CodeEQ(security.DefaultPopulation)).Exist(bare)
	if err != nil {
		t.Fatalf("population exist: %v", err)
	}
	if !popSeen {
		if _, err = db.Population.Create().SetCode(security.DefaultPopulation).SetName("users").SetPlane(security.PlaneBackend).Save(bare); err != nil {
			t.Fatalf("population: %v", err)
		}
	}
	tn, err := db.Tenant.Create().SetCode(liveTenant).SetName("e2e liveness").Save(bare)
	if err != nil {
		t.Fatalf("tenant: %v", err)
	}
	hash, err := utils.Hash("e2e-liveness-pass")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	u, err := db.User.Create().SetAccount(liveAccount).SetPasswordHash(hash).SetEnabled(true).Save(bare)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	if _, err = db.Membership.Create().SetUserID(u.ID).SetTenantID(liveTenant).SetPopulation(security.DefaultPopulation).Save(bare); err != nil {
		t.Fatalf("membership: %v", err)
	}

	userRP := usermod.NewUserRP(logger, pf)
	authzRP, err := authzmod.NewAuthzRP(logger, pf)
	if err != nil {
		t.Fatalf("authz rp: %v", err)
	}
	tenantRP := tenantmod.NewTenantRP(logger, pf)
	populationRP := populationmod.NewPopulationRP(logger, pf)
	tokenRP := authmod.NewTokenRP(logger, pf)
	providers, _ := e2eProviders(t, logger, cacheClient, tokenRP)
	uc, err := authmod.NewAuthUC(logger, pf, nil, providers, userRP, authzRP, tenantRP, populationRP, tokenRP)
	if err != nil {
		t.Fatalf("auth uc: %v", err)
	}

	token, _, _, err := uc.Login(bare, liveAccount, "e2e-liveness-pass", "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	// The mint embeds the version it saw, so the first verification rides
	// the comparison.
	if _, err = uc.AuthenticateSession(bare, token); err != nil {
		t.Fatalf("first verify: %v", err)
	}

	// A tenant flip moves the version inside the flip's own transaction, so
	// the very next verification pays the deep read and dies.
	s0 := stampRead(t, db)
	if _, err = db.Tenant.UpdateOneID(tn.ID).SetEnabled(false).Save(bare); err != nil {
		t.Fatalf("tenant disable: %v", err)
	}
	if s1 := stampRead(t, db); s1 <= s0 {
		t.Fatalf("tenant disable must bump: %d → %d", s0, s1)
	}
	if _, err = uc.AuthenticateSession(bare, token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled tenant: want unauthenticated, got %v", err)
	}

	// The gate is a check, not a revocation: the re-enable resurrects the
	// stored key, the deep read stamps the payload, and a second verify
	// rides the comparison to the same answer.
	if _, err = db.Tenant.UpdateOneID(tn.ID).SetEnabled(true).Save(bare); err != nil {
		t.Fatalf("tenant enable: %v", err)
	}
	if _, err = uc.AuthenticateSession(bare, token); err != nil {
		t.Fatalf("re-enabled tenant: session must resurrect, got %v", err)
	}
	if _, err = uc.AuthenticateSession(bare, token); err != nil {
		t.Fatalf("second verify after write-back: %v", err)
	}

	// The identity gate rides the same channel: a disable moves the
	// version and the next verification dies; the re-enable resurrects.
	s1 := stampRead(t, db)
	if _, err = db.User.UpdateOneID(u.ID).SetEnabled(false).Save(bare); err != nil {
		t.Fatalf("user disable: %v", err)
	}
	if s2 := stampRead(t, db); s2 <= s1 {
		t.Fatalf("user disable must bump: %d → %d", s1, s2)
	}
	if _, err = uc.AuthenticateSession(bare, token); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("disabled identity: want unauthenticated, got %v", err)
	}
	if _, err = db.User.UpdateOneID(u.ID).SetEnabled(true).Save(bare); err != nil {
		t.Fatalf("user enable: %v", err)
	}
	if _, err = uc.AuthenticateSession(bare, token); err != nil {
		t.Fatalf("re-enabled identity: session must resurrect, got %v", err)
	}
}
