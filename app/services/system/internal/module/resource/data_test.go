package resource

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"cyber-ecosystem/app/services/system/internal/conf"
)

// writeManifest drops one catalog file; these laws ride the real DirFS, not a
// fake — the json contract is the unit under test.
func writeManifest(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

const svcA = `[{"name":"A","full_name":"cyber.example.A","package":"cyber.example","methods":[{"name":"M"}]}]`
const svcB = `[{"name":"B","full_name":"cyber.example.B","package":"cyber.example","methods":[{"name":"M"}]}]`

func newRP(t *testing.T, dir string) ResourceRP {
	t.Helper()
	rp, err := NewResourceRP(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, &conf.Catalog{Dir: dir})
	if err != nil {
		t.Fatalf("new rp: %v", err)
	}
	return rp
}

// A service declared by two manifests is refused at boot.
func TestCatalogDuplicateServiceFails(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "a.json", svcA)
	writeManifest(t, dir, "b.json", svcA)

	if _, err := NewResourceRP(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, &conf.Catalog{Dir: dir}); err == nil {
		t.Fatal("duplicate service: loader must refuse")
	}
}

// An empty catalog is a misconfiguration — the grant tree has nothing to
// authorize against.
func TestCatalogEmptyFails(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "notes.txt", "not a manifest")

	if _, err := NewResourceRP(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, &conf.Catalog{Dir: dir}); err == nil {
		t.Fatal("empty catalog: loader must refuse")
	}
}

// The happy catalog loads sorted by full name; non-manifest files are
// ignored.
func TestCatalogLoadsSorted(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, "b.json", svcB)
	writeManifest(t, dir, "a.json", svcA)
	writeManifest(t, dir, "ignore.txt", "x")

	rp := newRP(t, dir)
	list, err := rp.ListResource(t.Context())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].FullName != "cyber.example.A" || list[1].FullName != "cyber.example.B" {
		t.Fatalf("list = %+v, want A then B", list)
	}
}
