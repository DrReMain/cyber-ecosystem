package e2e

// The database-only audit sink, live: a published event drains into an
// audit_log row carrying the event's true time (replays never skew the
// timeline) and a fresh event id. The constraint-loss branch is a
// defensive arm with no trigger path on this platform — every insert
// mints a new xid, and there is no redelivery consumer — so it stays
// unexercised by design; the unique index it guards is the schema's
// documented dedup contract.
// Gated: ENT_E2E=1 with the dev stack (postgres) up runs it.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/ent"
	entauditlog "cyber-ecosystem/app/services/agent/internal/ent/auditlog"
	auditmod "cyber-ecosystem/app/services/agent/internal/module/audit"
	"cyber-ecosystem/app/services/agent/internal/platform"
)

const aiTenant = "e2e-audit-insert"

func TestAuditInsertLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres + redis) up to run the audit-insert end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	bare := context.Background()

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: twEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: twEnvOr("DB_PASS", "postgres"), DbName: twEnvOr("DB_NAME", "agent"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler())
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanAuditInsert(t, db)
	t.Cleanup(func() {
		cleanAuditInsert(t, db)
		closeDB()
	})

	// Microsecond precision: the column keeps six digits, so the probe's
	// own clock value stays inside what a round-trip can carry.
	occurred := time.Date(2026, 11, 17, 1, 2, 3, 1000, time.UTC)
	rp, stop := auditmod.NewAuditRP(logger, pf)

	rp.Publish(bare, &kaudit.Event{
		TenantID: aiTenant, Actor: "e2e-audit-actor", Operation: "e2e audit insert",
		HTTPMethod: "GET", HTTPPath: "/e2e/audit", Status: 200, LatencyMs: 7, OccurredAt: occurred,
	})
	// stop drains the queue; calling it twice would close a closed channel.
	stop()

	row, rerr := db.AuditLog.Query().Where(
		entauditlog.TenantEQ(aiTenant), entauditlog.OperationEQ("e2e audit insert")).Only(bare)
	if rerr != nil {
		t.Fatalf("audit row: %v", rerr)
	}
	if row.EventID == "" {
		t.Fatalf("event id: %q", row.EventID)
	}
	if !row.CreatedAt.Equal(occurred) {
		t.Fatalf("created at: %v, want the event's own %v", row.CreatedAt, occurred)
	}
	if row.Actor != "e2e-audit-actor" || row.Status != 200 {
		t.Fatalf("row shape: %+v", row)
	}
}

func cleanAuditInsert(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.AuditLog.Delete().Where(entauditlog.TenantEQ(aiTenant)).Exec(ctx); err != nil {
		t.Fatalf("clean audit rows: %v", err)
	}
}
