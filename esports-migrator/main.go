// Package main is the one-shot legacy importer for the esports cutover: it
// reads the old MySQL store and writes the new system (identity core) and
// esports (business) databases. It runs as a deploy-plane job — rehearsal
// against the unmasked production copy first (user ruling 2026-11-02), then the cutover window —
// never inside a live service process.
//
// It is deliberately its own Go module with zero imports from the platform
// tree: a one-shot tool must not live inside a service's internal reach,
// and the service's ent tree is unimportable by construction anyway. Both
// targets are written as raw SQL whose column lists are pinned to the
// services' Atlas migrations; the rehearsal's row-count and per-field
// sampling pass is the drift check for that seam. The tool owns its own
// bookkeeping table (migration_map) — migration evidence, not product
// schema — and after cutover this whole directory is deleted in one
// commit.
//
// The run is phase-structured (--phase) and idempotent: identity writes
// land behind ON CONFLICT DO NOTHING on the system side and behind
// migration_map rows on the esports side, so a failed run is re-run, not
// rolled back.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type environment struct {
	src    *sql.DB
	sys    *sql.DB
	esp    *sql.DB
	dryRun bool
	sample int
}

// ensureTargets opens the write targets on first use: read-only phases
// (guard, dry-runs) must run without target credentials.
func (e *environment) ensureTargets() error {
	if err := e.ensureSystem(); err != nil {
		return err
	}
	return e.ensureEsports()
}

func (e *environment) ensureSystem() error {
	if e.sys != nil {
		return nil
	}
	sysDSN, err := requiredEnv("APP_MIG_SYSTEM_DSN")
	if err != nil {
		return err
	}
	e.sys, err = openPostgres("system", sysDSN)
	return err
}

// ensureEsports opens only the esports target: the entitlement phase writes
// grants and codes but never touches the identity core.
func (e *environment) ensureEsports() error {
	if e.esp != nil {
		return nil
	}
	espDSN, err := requiredEnv("APP_MIG_ESPORTS_DSN")
	if err != nil {
		return err
	}
	e.esp, err = openPostgres("esports", espDSN)
	return err
}

func main() {
	phaseFlag := flag.String("phase", "all", "comma-separated phases to run (guard,identity,... or all)")
	dryRun := flag.Bool("dry-run", false, "read the source and print the plan census without writing anything")
	sample := flag.Int("sample", 200, "per-field sampling size for the reconcile phase")
	flag.Parse()

	env := &environment{dryRun: *dryRun, sample: *sample}

	srcDSN, err := requiredEnv("APP_MIG_SOURCE_DSN")
	if err != nil {
		fatal(err)
	}
	env.src, err = openSource(srcDSN)
	if err != nil {
		fatal(fmt.Errorf("open source: %w", err))
	}
	defer func() { _ = env.src.Close() }()

	names, err := expandPhases(*phaseFlag)
	if err != nil {
		fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	for _, name := range names {
		fmt.Printf("\n=== phase %s ===\n", name)
		if err := phases[name](ctx, env); err != nil {
			fatal(fmt.Errorf("phase %s: %w", name, err))
		}
	}
	fmt.Println("\nmigrator: done")
	if env.esp != nil {
		_ = env.esp.Close()
	}
	if env.sys != nil {
		_ = env.sys.Close()
	}
}

// phases lists every planned phase in migration order. Phases beyond the
// current cut are registered as explicit stubs so naming one is a clear
// error instead of a silent skip.
var phases = map[string]func(context.Context, *environment) error{
	"guard":       runGuard,
	"identity":    runIdentity,
	"entitlement": runEntitlement,
	"behavior":    runBehavior,
	"snapshot":    stubPhase("snapshot"),
	"recompute":   stubPhase("recompute"),
	"reconcile":   runReconcile,
}

// order mirrors the frozen migration order: identity first (it produces the
// id map everything else joins), then entitlements, then behavior, then
// report snapshots, with recompute and reconcile closing.
var order = []string{"guard", "identity", "entitlement", "behavior", "snapshot", "recompute", "reconcile"}

func stubPhase(name string) func(context.Context, *environment) error {
	return func(context.Context, *environment) error {
		return fmt.Errorf("phase %q lands in a later MIG1 cut", name)
	}
}

func expandPhases(flagValue string) ([]string, error) {
	if flagValue == "all" {
		return order[:4], nil // the phases this cut implements
	}
	var out []string
	for _, raw := range strings.Split(flagValue, ",") {
		name := strings.TrimSpace(raw)
		if _, ok := phases[name]; !ok {
			return nil, fmt.Errorf("unknown phase %q (known: %s)", name, strings.Join(order, ","))
		}
		out = append(out, name)
	}
	return out, nil
}

func requiredEnv(name string) (string, error) {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return "", fmt.Errorf("environment variable %s is required", name)
	}
	return v, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "migrator:", err)
	os.Exit(1)
}
