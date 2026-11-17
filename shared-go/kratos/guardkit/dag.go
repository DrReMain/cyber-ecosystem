package guardkit

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

// GoFile is one parsed handwritten file: its slash path relative to the
// walked root, its imports, and (full-parse walks only) the string
// literals, HookRegistry references, and *Service struct declarations the
// composition guards read.
type GoFile struct {
	Rel            string
	Imports        []string
	Literals       []string
	HookRefs       int
	ServiceStructs []string
}

// Files walks root collecting parsed files. When ruleDir is non-empty
// only that directory subtree is parsed, imports-only (the DAG check);
// otherwise every handwritten file (ent generated code excluded, ent/schema
// included) is fully parsed to surface string literals and contract refs.
func Files(t *testing.T, root, ruleDir string, importsOnly bool) []GoFile {
	t.Helper()
	root = filepath.Clean(root)
	var out []GoFile
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil // the root's own entry is not a hidden dir to skip
		}
		rel := filepath.ToSlash(strings.TrimPrefix(filepath.ToSlash(p), filepath.ToSlash(root)+"/"))
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
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
		f, err := parser.ParseFile(fset, p, nil, mode)
		if err != nil {
			t.Fatalf("parse %s: %v", rel, err)
		}
		gf := GoFile{Rel: rel}
		for _, imp := range f.Imports {
			v, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			gf.Imports = append(gf.Imports, v)
		}
		if !importsOnly {
			ast.Inspect(f, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.BasicLit:
					if n.Kind == token.STRING {
						if v, err := strconv.Unquote(n.Value); err == nil {
							gf.Literals = append(gf.Literals, v)
						}
					}
				case *ast.Ident:
					if n.Name == "HookRegistry" {
						gf.HookRefs++
					}
				case *ast.TypeSpec:
					if _, isStruct := n.Type.(*ast.StructType); isStruct && strings.HasSuffix(n.Name.Name, "Service") {
						gf.ServiceStructs = append(gf.ServiceStructs, n.Name.Name)
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

// AssertModuleDAG enforces the service's banned-imports table: the module
// dependency DAG as code. Direct imports only — a transitive leak requires
// importing the middleman, which the table catches at that hop. A rule
// whose directory exists but matches no files fails as stale.
func AssertModuleDAG(t *testing.T, internalRoot string, banned map[string][]string) {
	t.Helper()
	for dir, bans := range banned {
		files := Files(t, internalRoot, dir, true)
		if len(files) == 0 {
			if _, statErr := os.Stat(filepath.Join(internalRoot, dir)); statErr == nil {
				// The directory exists but matched nothing: renamed files or
				// an empty package. A not-yet-created module is expected to
				// match nothing.
				t.Errorf("rule %q matched no files — stale directory name?", dir)
			}
		}
		for _, f := range files {
			for _, imp := range f.Imports {
				for _, ban := range bans {
					if strings.Contains(imp, "/"+ban) {
						t.Errorf("%s imports %q — %s must not depend on %s (module DAG)",
							f.Rel, imp, dir, ban)
					}
				}
			}
		}
	}
}

// AssertCompositionBoundaries pins the composition layering: hand-written
// cmd/app assembly reaches business modules only through wire provider
// sets, and injected app contracts (HookRegistry) land in module biz
// constructors — a repo's phase machinery is exposed as port methods and
// registered by its UC; data/client adapters never receive the registry.
func AssertCompositionBoundaries(t *testing.T, internalRoot string) {
	t.Helper()
	for _, f := range Files(t, filepath.Join(internalRoot, "..", "cmd", "app"), "", true) {
		if f.Rel == "wire.go" || f.Rel == "wire_gen.go" {
			continue
		}
		for _, imp := range f.Imports {
			if strings.Contains(imp, "/internal/module/") {
				t.Errorf("%s imports %q — cmd/app hand-written assembly wires via bootstrap/server only; modules enter through wire provider sets", f.Rel, imp)
			}
		}
	}
	for _, f := range Files(t, internalRoot, "module", false) {
		if f.HookRefs == 0 {
			continue
		}
		if !strings.HasSuffix(f.Rel, "/biz.go") {
			t.Errorf("%s references HookRegistry — injected app contracts land in biz constructors only", f.Rel)
		}
	}
}

// AssertModuleSingleService pins one module = one transport face: at most
// one *Service struct per module dir. Two faces in one package mean two
// aggregates sharing a module — the cure is the module split, never a
// second service struct.
func AssertModuleSingleService(t *testing.T, internalRoot string) {
	t.Helper()
	counts := map[string][]string{}
	for _, f := range Files(t, internalRoot, "module", false) {
		if len(f.ServiceStructs) == 0 {
			continue
		}
		mod, _, _ := strings.Cut(strings.TrimPrefix(f.Rel, "module/"), "/")
		counts[mod] = append(counts[mod], f.ServiceStructs...)
	}
	for mod, names := range counts {
		if len(names) > 1 {
			t.Errorf("module/%s declares %d service structs (%s) — one module one face; split the module",
				mod, len(names), strings.Join(names, ", "))
		}
	}
}

// AssertLiteralSingleSource pins contract strings to the single file
// allowed to declare them (owners) and fails on banned literals spelled
// raw anywhere (banned) — the owning package's constant is the only
// spelling. Services without their own contract vocabulary skip this
// check.
func AssertLiteralSingleSource(t *testing.T, internalRoot string, owners map[string]string, banned []string) {
	t.Helper()
	files := Files(t, internalRoot, "", false)
	if len(files) == 0 {
		t.Fatal("walk matched no files — guard root is broken")
	}
	for _, f := range files {
		for _, lit := range f.Literals {
			if owner, ok := owners[lit]; ok {
				if f.Rel != owner {
					t.Errorf("%s declares %q — owned by %s (seam 3: contract strings live once)",
						f.Rel, lit, owner)
				}
				continue
			}
			for _, ban := range banned {
				if lit == ban {
					t.Errorf("%s spells %q as a raw literal — use the owning package's constant",
						f.Rel, lit)
				}
			}
		}
	}
}
