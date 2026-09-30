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

// Static guards for the dependency rules iam.md §4.3 spells out as prose:
// the module DAG, and the shared-contract seams that must hold until the
// multi-service split (§4.4). Seam 2 (compile() stays a pure function) is a
// review-enforced property a file walk cannot see; seams 1 and 3 are checked
// here, plus single-sourcing of the ABAC vocabulary.

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
		files := goFiles(t, dir, true)
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
						t.Errorf("%s imports %q — %s must not depend on %s (iam.md DAG)",
							f.rel, imp, dir, ban)
					}
				}
			}
		}
	}
}

func TestSharedContractSingleSource(t *testing.T) {
	files := goFiles(t, "", false)
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
}

// goFiles walks internal/ collecting parsed files. When ruleDir is non-empty
// only that directory subtree is parsed, imports-only (the DAG check);
// otherwise every handwritten file (ent generated code excluded, ent/schema
// included) is fully parsed to surface string literals.
func goFiles(t *testing.T, ruleDir string, importsOnly bool) []goFile {
	t.Helper()
	root := filepath.Clean("..") // the shared package sits directly under internal/
	var out []goFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil // the root's own entry ("..") is not a hidden dir to skip
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
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if v, err := strconv.Unquote(lit.Value); err == nil {
						gf.literals = append(gf.literals, v)
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
