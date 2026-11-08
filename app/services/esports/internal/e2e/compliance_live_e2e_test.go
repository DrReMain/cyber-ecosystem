package e2e

// The compliance knobs' live probe: the absent row answering the shipped
// defaults, each setter moving only its own column (concurrent setters on
// different knobs never clobber each other), the prompt override landing
// and holding while its sibling stays unset, the age line's bounds
// refusal, and the audit-join column carrying the acting admin. Gated:
// ENT_E2E=1 with the dev database up; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/module/compliance"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

const complianceAdmin = "compliance-e2e" // the audit-join value the setters stamp

func TestComplianceLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the compliance knobs end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: entEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: entEnvOr("DB_PASS", "postgres"), DbName: entEnvOr("DB_NAME", "esports"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	t.Cleanup(func() {
		cleanComplianceE2E(t, db)
		closeDB()
	})
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanComplianceE2E(t, db)

	rp := compliance.NewComplianceRP(logger, pf)
	uc := compliance.NewComplianceUC(logger, pf, rp)

	// The absent row answers the shipped defaults — every knob unset, no
	// prompt override.
	base, err := uc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if base.BirthCollect != compliance.DefaultBirthCollect || base.PayAgeLine != compliance.DefaultPayAgeLine {
		t.Fatalf("base: %+v, want shipped defaults", base)
	}
	if base.MinorPrompt != nil || base.MissingBirthPrompt != nil {
		t.Fatalf("base prompts: %v/%v, want nil overrides", base.MinorPrompt, base.MissingBirthPrompt)
	}

	// No session, no write — updated_by is an audit fact, never blank.
	if _, err := uc.SetBirthCollection(context.Background(), false); err == nil || !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("anonymous set: err=%v, want unauthenticated", err)
	}

	// Each setter moves only its own column: the birth flip leaves the
	// age line and both prompts exactly where they were.
	cfg, err := uc.SetBirthCollection(entSubject(complianceAdmin), false)
	if err != nil {
		t.Fatalf("set birth: %v", err)
	}
	if cfg.BirthCollect || cfg.PayAgeLine != compliance.DefaultPayAgeLine || cfg.MinorPrompt != nil || cfg.MissingBirthPrompt != nil {
		t.Fatalf("after birth flip: %+v, want only birth moved", cfg)
	}
	cfg, err = uc.SetPayAgeLine(entSubject(complianceAdmin), 16)
	if err != nil {
		t.Fatalf("set age: %v", err)
	}
	if cfg.BirthCollect || cfg.PayAgeLine != 16 || cfg.MinorPrompt != nil || cfg.MissingBirthPrompt != nil {
		t.Fatalf("after age set: %+v, want only age moved", cfg)
	}

	// A prompt edit lands without disturbing its sibling — nil carries
	// "leave that one untouched" on both calls.
	cfg, err = uc.SetCompliancePrompts(entSubject(complianceAdmin), utils.Ptr("e2e minor copy"), nil)
	if err != nil {
		t.Fatalf("set minor prompt: %v", err)
	}
	if cfg.MinorPrompt == nil || *cfg.MinorPrompt != "e2e minor copy" || cfg.MissingBirthPrompt != nil || cfg.BirthCollect || cfg.PayAgeLine != 16 {
		t.Fatalf("after minor prompt: %+v, want only minor set", cfg)
	}
	cfg, err = uc.SetCompliancePrompts(entSubject(complianceAdmin), nil, utils.Ptr("e2e missing copy"))
	if err != nil {
		t.Fatalf("set missing prompt: %v", err)
	}
	if cfg.MissingBirthPrompt == nil || *cfg.MissingBirthPrompt != "e2e missing copy" || cfg.MinorPrompt == nil || *cfg.MinorPrompt != "e2e minor copy" {
		t.Fatalf("after missing prompt: %+v, want minor held and missing set", cfg)
	}

	// The age line's bounds are the same refusal a direct caller gets —
	// the wire enforces them, the UC does too.
	for _, bad := range []int32{0, -1, 121} {
		if _, err := uc.SetPayAgeLine(entSubject(complianceAdmin), bad); err == nil || !errorspb.IsGeneralErrorInvalidArgument(err) {
			t.Fatalf("age %d: err=%v, want invalid argument", bad, err)
		}
	}

	// The audit-join column carries the acting admin, not a blank.
	row, err := db.ComplianceConfig.Query().Only(context.Background())
	if err != nil {
		t.Fatalf("row: %v", err)
	}
	if row.UpdatedBy != complianceAdmin {
		t.Fatalf("updated_by: %q, want %q", row.UpdatedBy, complianceAdmin)
	}
}

func cleanComplianceE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	if _, err := db.ComplianceConfig.Delete().Exec(context.Background()); err != nil {
		t.Fatalf("clean compliance rows: %v", err)
	}
}
