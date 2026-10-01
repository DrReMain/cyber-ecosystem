package shared

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Static guards for the dependency rules kratos CONVENTIONS §2/§3 spell out
// as prose: the module DAG, the shared-contract seams that must hold until
// the multi-service split, and the composition layering (cmd/app wires
// through bootstrap; shared stays a contracts-only kernel; injected app
// contracts land in module biz constructors). A pure-function property like
// compile() is review-enforced — a file walk cannot see it; the rest is
// checked here, plus single-sourcing of the ABAC vocabulary.

// bannedImports maps a package directory (relative to internal/, prefix
// match) to import markers it must not contain. Direct imports only: a
// transitive leak requires importing the middleman, which these rules catch
// at that hop.
var bannedImports = map[string][]string{
	// The engine reads ent tables directly and never sibling modules.
	"module/authz": {
		"module/role", "module/policy", "module/dept", "module/user",
		"module/auth", "module/resource", "module/transfer",
	},
	// role may read the authz port (diagnostics trio); the others must not.
	"module/policy": {"module/authz"},
	"module/dept":   {"module/authz"},
	"module/user":   {"module/authz", "module/auth"},
	"module/role":   {"module/user"},
	// Seam 1: the session-auth closure stays free of ent so it can move to
	// shared-go at the split as-is.
	"module/auth": {"internal/ent"},
	// CONVENTIONS §3: shared → platform → ent closes a cycle.
	"ent/schema": {"internal/shared"},
	// Bootstrap sits below business modules.
	"bootstrap": {"module/"},
	// The kernel stays ent-free (Paginate is generic by design).
	"shared": {"internal/ent"},
}

// literalOwners pins contract strings to the single file allowed to declare
// them; bannedLiterals must not appear in service code at all — the kauthz
// constant is the only spelling.
var (
	literalOwners = map[string]string{
		"authz:policy:ver":     "shared/policy_notify.go",
		"authz:policy:changed": "shared/policy_notify.go",
		"auth:web:revoked:":    "shared/session_revoke.go",
	}
	bannedLiterals = []string{"time_window", "calendar", "time.now"}
)

func TestModuleDependencyDAG(t *testing.T) {
	for dir, bans := range bannedImports {
		files := goFiles(t, "..", dir, true)
		if len(files) == 0 {
			if _, statErr := os.Stat(filepath.Join("..", dir)); statErr == nil {
				// The directory exists but matched nothing: renamed files or
				// an empty package. A not-yet-created module (e.g. policy
				// before S2) is expected to match nothing for now.
				t.Errorf("rule %q matched no files — stale directory name?", dir)
			}
		}
		for _, f := range files {
			for _, imp := range f.imports {
				for _, ban := range bans {
					if strings.Contains(imp, "/"+ban) {
						t.Errorf("%s imports %q — %s must not depend on %s (CONVENTIONS §3 DAG)",
							f.rel, imp, dir, ban)
					}
				}
			}
		}
	}
}

// TestCompositionBoundaries pins the composition layering: hand-written
// cmd/app assembly reaches business modules only through wire provider
// sets, and injected app contracts (HookRegistry) land in module biz
// constructors — a repo's phase machinery is exposed as port methods and
// registered by its UC; data/client adapters never receive the registry.
func TestCompositionBoundaries(t *testing.T) {
	for _, f := range goFiles(t, filepath.Clean("../../cmd/app"), "", true) {
		if f.rel == "wire.go" || f.rel == "wire_gen.go" {
			continue
		}
		for _, imp := range f.imports {
			if strings.Contains(imp, "/internal/module/") {
				t.Errorf("%s imports %q — cmd/app hand-written assembly wires via bootstrap/server only; modules enter through wire provider sets", f.rel, imp)
			}
		}
	}
	for _, f := range goFiles(t, "..", "module", false) {
		if f.hookRefs == 0 {
			continue
		}
		if !strings.HasSuffix(f.rel, "/biz.go") {
			t.Errorf("%s references HookRegistry — injected app contracts land in biz constructors only", f.rel)
		}
	}
}

func TestSharedContractSingleSource(t *testing.T) {
	files := goFiles(t, "..", "", false)
	if len(files) == 0 {
		t.Fatal("walk matched no files — guard root is broken")
	}
	for _, f := range files {
		for _, lit := range f.literals {
			if owner, ok := literalOwners[lit]; ok {
				if f.rel != owner {
					t.Errorf("%s declares %q — owned by %s (seam 3: contract strings live once)",
						f.rel, lit, owner)
				}
				continue
			}
			for _, ban := range bannedLiterals {
				if lit == ban {
					t.Errorf("%s spells %q as a raw literal — use the kauthz constant",
						f.rel, lit)
				}
			}
		}
	}
}

type goFile struct {
	rel      string
	imports  []string
	literals []string
	hookRefs int
}

// goFiles walks root collecting parsed files. When ruleDir is non-empty
// only that directory subtree is parsed, imports-only (the DAG check);
// otherwise every handwritten file (ent generated code excluded, ent/schema
// included) is fully parsed to surface string literals and contract refs.
func goFiles(t *testing.T, root, ruleDir string, importsOnly bool) []goFile {
	t.Helper()
	root = filepath.Clean(root)
	var out []goFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil // the root's own entry is not a hidden dir to skip
		}
		rel := filepath.ToSlash(strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(root)+"/"))
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if ruleDir != "" && (rel != ruleDir && !strings.HasPrefix(rel, ruleDir+"/")) {
			return nil
		}
		if ruleDir == "" && strings.HasPrefix(rel, "ent/") && !strings.HasPrefix(rel, "ent/schema/") {
			return nil // generated: imports never drift, literals never authored
		}
		mode := parser.ImportsOnly
		if !importsOnly {
			mode = parser.SkipObjectResolution
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, mode)
		if err != nil {
			t.Fatalf("parse %s: %v", rel, err)
		}
		gf := goFile{rel: rel}
		for _, imp := range f.Imports {
			v, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			gf.imports = append(gf.imports, v)
		}
		if !importsOnly {
			ast.Inspect(f, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.BasicLit:
					if n.Kind == token.STRING {
						if v, err := strconv.Unquote(n.Value); err == nil {
							gf.literals = append(gf.literals, v)
						}
					}
				case *ast.Ident:
					if n.Name == "HookRegistry" {
						gf.hookRefs++
					}
				}
				return true
			})
		}
		out = append(out, gf)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}
