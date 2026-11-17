package guardkit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// GoFile is one parsed handwritten file: its slash path relative to the
// walked root, its imports, and (full-parse walks only) the string
// literals, HookRegistry references, *Service struct declarations, and
// wall-clock reads the composition and clock guards read.
type GoFile struct {
	Rel            string
	Imports        []string
	TimeAlias      string // local name of the "time" import (alias or "time"); "" when absent
	Literals       []string
	HookRefs       int
	ServiceStructs []string
	ClockReads     []ClockRead
}

// ClockRead is one call to the time package's read family. Timers and
// tickers (After, NewTimer, NewTicker, Tick) are pacing primitives, not
// clock reads, and are deliberately absent.
type ClockRead struct {
	Fn   string // Now, Until, Since — every spelling that yields a wall-clock value
	Line int    // 1-based line of the call
}

var nowReaders = map[string]bool{"Now": true, "Until": true, "Since": true}

// Files walks root collecting parsed files. When ruleDir is non-empty
// only that directory subtree is parsed, imports-only (the DAG check);
// otherwise every handwritten file (ent generated code excluded, ent/schema
// included) is fully parsed to surface string literals, contract refs, and
// clock reads. The walk is a pure function: parse failures return an error
// instead of failing a test, so the Find* detectors below stay assertable
// against testdata trees.
func Files(root, ruleDir string, importsOnly bool) ([]GoFile, error) {
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
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		gf := GoFile{Rel: rel}
		for _, imp := range f.Imports {
			v, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			gf.Imports = append(gf.Imports, v)
			if v == "time" {
				gf.TimeAlias = "time"
				if imp.Name != nil {
					gf.TimeAlias = imp.Name.Name
				}
			}
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
				case *ast.CallExpr:
					sel, ok := n.Fun.(*ast.SelectorExpr)
					if !ok {
						break
					}
					id, ok := sel.X.(*ast.Ident)
					if !ok || id.Name != gf.TimeAlias || gf.TimeAlias == "" {
						break
					}
					if nowReaders[sel.Sel.Name] {
						gf.ClockReads = append(gf.ClockReads, ClockRead{Fn: sel.Sel.Name, Line: fset.Position(n.Pos()).Line})
					}
				}
				return true
			})
		}
		out = append(out, gf)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	return out, nil
}

// FindModuleDAG reports the service's banned-imports violations: the module
// dependency DAG as code. Direct imports only — a transitive leak requires
// importing the middleman, which the table catches at that hop. A rule
// whose directory exists but matches no files is stale.
func FindModuleDAG(internalRoot string, banned map[string][]string) ([]Violation, error) {
	var out []Violation
	for dir, bans := range banned {
		files, err := Files(internalRoot, dir, true)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			if _, statErr := os.Stat(filepath.Join(internalRoot, dir)); statErr == nil {
				// The directory exists but matched nothing: renamed files or
				// an empty package. A not-yet-created module is expected to
				// match nothing.
				out = append(out, Violation{Pos: dir,
					Message: fmt.Sprintf("rule %q matched no files — stale directory name?", dir)})
			}
		}
		for _, f := range files {
			for _, imp := range f.Imports {
				for _, ban := range bans {
					if strings.Contains(imp, "/"+ban) {
						out = append(out, Violation{Pos: f.Rel,
							Message: fmt.Sprintf("%s imports %q — %s must not depend on %s (module DAG)", f.Rel, imp, dir, ban)})
					}
				}
			}
		}
	}
	return sortedViolations(out), nil
}

// FindCompositionBoundaries reports composition-layering violations:
// hand-written cmd/app assembly reaching business modules directly instead
// of through wire provider sets, and injected app contracts (HookRegistry)
// landing outside module biz constructors — a repo's phase machinery is
// exposed as port methods and registered by its UC; data/client adapters
// never receive the registry.
func FindCompositionBoundaries(internalRoot string) ([]Violation, error) {
	var out []Violation
	files, err := Files(filepath.Join(internalRoot, "..", "cmd", "app"), "", true)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.Rel == "wire.go" || f.Rel == "wire_gen.go" {
			continue
		}
		for _, imp := range f.Imports {
			if strings.Contains(imp, "/internal/module/") {
				out = append(out, Violation{Pos: f.Rel,
					Message: fmt.Sprintf("%s imports %q — cmd/app hand-written assembly wires via bootstrap/server only; modules enter through wire provider sets", f.Rel, imp)})
			}
		}
	}
	modFiles, err := Files(internalRoot, "module", false)
	if err != nil {
		return nil, err
	}
	for _, f := range modFiles {
		if f.HookRefs == 0 {
			continue
		}
		if !strings.HasSuffix(f.Rel, "/biz.go") {
			out = append(out, Violation{Pos: f.Rel,
				Message: fmt.Sprintf("%s references HookRegistry — injected app contracts land in biz constructors only", f.Rel)})
		}
	}
	return sortedViolations(out), nil
}

// FindModuleSingleService reports modules declaring more than one *Service
// struct: one module = one transport face. Two faces in one package mean
// two aggregates sharing a module — the cure is the module split, never a
// second service struct.
func FindModuleSingleService(internalRoot string) ([]Violation, error) {
	counts := map[string][]string{}
	files, err := Files(internalRoot, "module", false)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if len(f.ServiceStructs) == 0 {
			continue
		}
		mod, _, _ := strings.Cut(strings.TrimPrefix(f.Rel, "module/"), "/")
		counts[mod] = append(counts[mod], f.ServiceStructs...)
	}
	var out []Violation
	for mod, names := range counts {
		if len(names) > 1 {
			out = append(out, Violation{Pos: "module/" + mod,
				Message: fmt.Sprintf("module/%s declares %d service structs (%s) — one module one face; split the module", mod, len(names), strings.Join(names, ", "))})
		}
	}
	return sortedViolations(out), nil
}

// FindLiteralSingleSource reports contract strings spelled outside the
// single file allowed to declare them (owners) and banned literals spelled
// raw anywhere (banned) — the owning package's constant is the only
// spelling. Services without their own contract vocabulary skip this
// check.
func FindLiteralSingleSource(internalRoot string, owners map[string]string, banned []string) ([]Violation, error) {
	var out []Violation
	files, err := Files(internalRoot, "", false)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("walk matched no files — guard root %s is broken", internalRoot)
	}
	for _, f := range files {
		for _, lit := range f.Literals {
			if owner, ok := owners[lit]; ok {
				if f.Rel != owner {
					out = append(out, Violation{Pos: f.Rel,
						Message: fmt.Sprintf("%s declares %q — owned by %s (seam 3: contract strings live once)", f.Rel, lit, owner)})
				}
				continue
			}
			for _, ban := range banned {
				if lit == ban {
					out = append(out, Violation{Pos: f.Rel,
						Message: fmt.Sprintf("%s spells %q as a raw literal — use the owning package's constant", f.Rel, lit)})
				}
			}
		}
	}
	return sortedViolations(out), nil
}

// FindBizClockViolations reports every bare wall-clock read in a module's
// handwritten files: any call to the time package's read family (Now,
// Until, Since — aliases covered), whatever file of the module package it
// lands in. The clock-port law governs the biz layer, and the biz layer is
// not a file name: a module's domain logic lives in any handwritten file of
// its package (attrs.go, policies.go, ...), so data.go and service.go — the
// RP and transport faces, where cache stamps and metric timing live — are
// the only exclusions. A read on a line carrying an inline // wall-clock:
// note is a declared presentation-only read.
func FindBizClockViolations(internalRoot string) ([]Violation, error) {
	files, err := Files(internalRoot, "module", false)
	if err != nil {
		return nil, err
	}
	fsys := os.DirFS(internalRoot)
	var out []Violation
	for _, f := range files {
		if base := path.Base(f.Rel); base == "data.go" || base == "service.go" {
			continue
		}
		if len(f.ClockReads) == 0 {
			continue
		}
		raw, err := fs.ReadFile(fsys, f.Rel)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(raw), "\n")
		for _, r := range f.ClockReads {
			if r.Line-1 < len(lines) && strings.Contains(lines[r.Line-1], wallClockNote) {
				continue // declared presentation-only read
			}
			out = append(out, Violation{Pos: fmt.Sprintf("%s:%d", f.Rel, r.Line),
				Message: fmt.Sprintf("%s:%d reads the wall clock bare (time.%s) — route it through the UC's clock field, or mark the line // wall-clock: with the reason it is presentation-only", f.Rel, r.Line, r.Fn)})
		}
	}
	return sortedViolations(out), nil
}
