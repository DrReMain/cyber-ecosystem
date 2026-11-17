package e2e

// The Postgres transaction-poisoning law, live: a failed statement poisons
// every later statement in its transaction (SQLSTATE 25P02 — Postgres has no
// statement-level rollback), which is exactly the class an in-process fake
// repo can never surface. The probe pins the fact itself: after one failed
// statement, the next one in the same transaction reports the poisoned
// state, and only rollback clears it.
//
// Gated: ENT_E2E=1 with the dev stack (postgres) up runs it.

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/platform"
)

func TestPostgresPoisoningLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres) up to run the poisoning probe")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: tenEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: tenEnvOr("DB_PASS", "postgres"), DbName: tenEnvOr("DB_NAME", "system"),
		MaxOpenConns: 2,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	defer closeDB()

	ctx := context.Background()
	tx, err := db.Tx(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	// The poisoning statement: division by zero errors server-side.
	if _, err := tx.ExecContext(ctx, "SELECT 1/0"); err == nil {
		t.Fatal("the poisoning statement must fail (1/0)")
	}

	// The poisoned successor: a perfectly valid statement now reports the
	// aborted-transaction state — no statement-level rollback exists.
	_, err = tx.ExecContext(ctx, "SELECT 1")
	if err == nil {
		t.Fatal("a valid statement after a failed one must report the poisoned transaction")
	}
	if !strings.Contains(err.Error(), "25P02") && !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("expected the poisoned-transaction shape (25P02 / aborted), got %v", err)
	}

	// Only rollback clears it; the connection returns to the pool healthy.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
		t.Fatalf("after rollback the connection must be healthy, got %v", err)
	}
}
