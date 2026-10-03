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

// Static guards for the dependency rules kratos CONVENTIONS spell out as
// prose: the module DAG and the composition layering (cmd/app wires through
// bootstrap; shared stays a contracts-only kernel; injected app contracts
// land in module biz constructors).

// bannedImports maps a package directory (relative to internal/, prefix
// match) to import markers it must not contain. Direct imports only: a
// transitive leak requires importing the middleman, which these rules catch
// at that hop.
var bannedImports = map[string][]string{
	// The engine floor is module-free infrastructure and never reads
	// business aggregates.
	"module/authz": {"module/agentconfig", "module/agentconfigadmin"},
	// The operator face is a strippable unit (its proto lives in its own
	// file); the business module must survive its deletion untouched.
	"module/agentconfig": {"module/agentconfigadmin"},
	// ACL adapters implement biz-declared ports (importing the declaring
	// module for the interface is the designed direction; the reverse is the
	// leak), and stay ent-free so the closure can move to shared-go as-is.
	"client": {"internal/ent"},
	// CONVENTIONS §3: shared → platform → ent closes a cycle.
	"ent/schema": {"internal/shared"},
	// Bootstrap sits below business modules.
	"bootstrap": {"module/"},
	// The kernel stays ent-free.
	"shared": {"internal/ent"},
}

func TestModuleDependencyDAG(t *testing.T) {
	for dir, bans := range bannedImports {
		files := goFiles(t, "..", dir, true)
		if len(files) == 0 {
			if _, statErr := os.Stat(filepath.Join("..", dir)); statErr == nil {
				// The directory exists but matched nothing: renamed files or
				// an empty package.
				t.Errorf("rule %q matched no files — stale directory name?", dir)
			}
		}
		for _, f := range files {
			for _, imp := range f.imports {
				for _, ban := range bans {
					if strings.Contains(imp, "/"+ban) {
						t.Errorf("%s imports %q — %s must not depend on %s (CONVENTIONS DAG)",
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

type goFile struct {
	rel      string
	imports  []string
	hookRefs int
}

// goFiles walks root collecting parsed files. When ruleDir is non-empty
// only that directory subtree is parsed, imports-only (the DAG check);
// otherwise every handwritten file (ent generated code excluded, ent/schema
// included) is parsed to surface contract refs.
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
			return nil // generated: imports never drift, refs never authored
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
				if id, ok := n.(*ast.Ident); ok && id.Name == "HookRegistry" {
					gf.hookRefs++
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
