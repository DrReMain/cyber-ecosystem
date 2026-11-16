package resource

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// The catalog loader's laws, on an in-memory FS: manifests load sorted by
// full name, a service declared twice is refused (the catalog must have
// one home per service), an empty or broken catalog refuses loudly —
// silence here would render every permission screen blank — and a stamp
// change on disk hot-reloads without a restart.

func newResourceRP(fs fstest.MapFS) *resourceRP {
	return &resourceRP{
		RP:       shared.NewRP(slog.New(slog.NewTextHandler(io.Discard, nil)), nil),
		fsys:     fs,
		dirLabel: "mem",
	}
}

// loadedResourceRP mirrors the production constructor: load runs once at
// build time and seeds the stamp, so later ListResource calls compare
// against a real stamp instead of the zero value.
func loadedResourceRP(t *testing.T, fs fstest.MapFS) *resourceRP {
	t.Helper()
	rp := newResourceRP(fs)
	if _, err := rp.load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	return rp
}

func manifest(full string, methods ...string) string {
	ops := make([]string, len(methods))
	for i, m := range methods {
		ops[i] = fmt.Sprintf(`{"name":%q}`, m)
	}
	return fmt.Sprintf(`[{"name":"%s","full_name":"%s","methods":[%s]}]`,
		full[strings.LastIndex(full, ".")+1:], full, strings.Join(ops, ","))
}

// Loads the manifests and answers them sorted by full name regardless of
// file naming or declaration order.
func TestListResourceLoadsSorted(t *testing.T) {
	rp := loadedResourceRP(t, fstest.MapFS{
		"z-first.json":  &fstest.MapFile{Data: []byte(manifest("cyber.zzz.Svc"))},
		"a.json":        &fstest.MapFile{Data: []byte(manifest("cyber.aaa.Svc", "One", "Two"))},
		"notes.txt":     &fstest.MapFile{Data: []byte("ignored")},
		"nested/b.json": &fstest.MapFile{Data: []byte("directories are not manifests")},
	})

	services, err := rp.ListResource(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(services) != 2 || services[0].FullName != "cyber.aaa.Svc" || services[1].FullName != "cyber.zzz.Svc" {
		t.Fatalf("services: %+v", services)
	}
	if len(services[0].Methods) != 2 {
		t.Fatalf("methods: %+v", services[0].Methods)
	}
}

// One service, two files: the catalog refuses the double declaration —
// an operator who split a catalog wrongly must hear about it, not get a
// silently arbitrary winner.
func TestListResourceRejectsDuplicateDeclaration(t *testing.T) {
	rp := newResourceRP(fstest.MapFS{
		"a.json": &fstest.MapFile{Data: []byte(manifest("cyber.dup.Svc"))},
		"b.json": &fstest.MapFile{Data: []byte(manifest("cyber.dup.Svc"))},
	})
	if _, err := rp.ListResource(context.Background()); err == nil {
		t.Fatal("duplicate declaration: want refusal")
	}
}

// A catalog with no manifests is a wiring error, not an empty answer —
// it refuses at construction, the production constructor's own gate.
func TestListResourceRejectsEmptyCatalog(t *testing.T) {
	rp := newResourceRP(fstest.MapFS{})
	if _, err := rp.load(); err == nil {
		t.Fatal("empty catalog: want refusal at load")
	}
}

// A broken manifest refuses with the file named — the operator's way in.
func TestListResourceRejectsBrokenJSON(t *testing.T) {
	rp := newResourceRP(fstest.MapFS{
		"a.json": &fstest.MapFile{Data: []byte("{not json")},
	})
	if _, err := rp.ListResource(context.Background()); err == nil {
		t.Fatal("broken json: want refusal")
	}
}

// The stamp watch: an unchanged directory serves the cached answer, and
// a manifest landing on disk between calls is picked up without a
// restart — permission screens follow deploys.
func TestListResourceHotReloadsOnStampChange(t *testing.T) {
	fs := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(manifest("cyber.aaa.Svc"))}}
	rp := loadedResourceRP(t, fs)
	ctx := context.Background()

	first, err := rp.ListResource(ctx)
	if err != nil || len(first) != 1 {
		t.Fatalf("first list: %d (%v)", len(first), err)
	}
	again, err := rp.ListResource(ctx)
	if err != nil || len(again) != 1 {
		t.Fatalf("unchanged list: %d (%v)", len(again), err)
	}

	fs["b.json"] = &fstest.MapFile{Data: []byte(manifest("cyber.bbb.Svc"))}
	reloaded, err := rp.ListResource(ctx)
	if err != nil || len(reloaded) != 2 {
		t.Fatalf("reloaded list: %d (%v)", len(reloaded), err)
	}
}
