// Package guardkit holds the shared machinery of the per-service convention
// guards: the file walks, the violation detectors, and the assert helpers
// every service's internal/guards package drives with its own topology
// tables (layer dirs, rung gates, module DAG). A service's guard files stay
// thin — data plus one-line tests — so a new guard lands here once and every
// service (and every fork) inherits it.
package guardkit

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var (
	dividerRE   = regexp.MustCompile(`^// [A-Za-z][A-Za-z0-9 /]* -{8,}\s*$`)
	directiveRE = regexp.MustCompile(`^// ?[a-z]+:`)
	declRE      = regexp.MustCompile(`^(type|func|const|var)\b`)
	kratosCode  = regexp.MustCompile(`errors\.Is(NotFound|InvalidArgument|Unauthorized|Forbidden|Conflict|BadRequest|InternalServerError|FailedPrecondition)\(`)
	jsonCall    = regexp.MustCompile(`json\.(Marshal|Unmarshal|NewEncoder|NewDecoder)\(`)

	bareClockRE   = regexp.MustCompile(`time\.Now\(\)`)
	wallClockNote = "// wall-clock:"
)

// LayerFiles walks the service's internal/ layer dirs and returns every
// handwritten .go file's content keyed by its slash path relative to the
// walked layer (server.go, module/auth/biz.go, ...). A dir the service
// lacks is skipped — layer sets differ across services and forks.
func LayerFiles(t *testing.T, internalRoot string, dirs []string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	for _, dir := range dirs {
		root := filepath.Join(internalRoot, dir)
		if _, statErr := os.Stat(root); statErr != nil {
			continue
		}
		// DirFS confines reads to the walked root (it rejects .. and
		// absolute paths), so the variable-path read carries no traversal
		// surface — the property G304 asks for, held structurally.
		fsys := os.DirFS(root)
		err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			data, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}
			out[p] = string(data)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	return out
}

// AssertNoDeclarationComments fails on the two illegal comment shapes: a
// prose block above the imports (neither divider nor directive) and a
// comment sitting directly on a declaration.
func AssertNoDeclarationComments(t *testing.T, files map[string]string) {
	t.Helper()
	for f, src := range files {
		lines := strings.Split(src, "\n")
		if i, ok := headerComment(lines); ok {
			t.Errorf("%s:%d file-header comment — neither legal kind lives above the imports: %q", f, i+1, strings.TrimSpace(lines[i]))
		}
		for i, line := range lines {
			s := strings.TrimSpace(line)
			if !strings.HasPrefix(s, "//") || dividerRE.MatchString(s) || directiveRE.MatchString(s) {
				continue
			}
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j < len(lines) && declRE.MatchString(strings.TrimSpace(lines[j])) {
				t.Errorf("%s:%d declaration comment: %q", f, i+1, s)
			}
		}
	}
}

// headerComment finds a comment line in the file-header position — after
// the package clause, before the first import or declaration. The comment
// law's two kinds are dividers and in-body why; a block floating above the
// imports is the third shape, and it escapes the declaration signature (the
// next non-blank line opens import, not a declaration).
func headerComment(lines []string) (int, bool) {
	seenPackage := false
	for i, line := range lines {
		s := strings.TrimSpace(line)
		if !seenPackage {
			seenPackage = strings.HasPrefix(s, "package ")
			continue
		}
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "//") && !dividerRE.MatchString(s) && !directiveRE.MatchString(s) {
			return i, true
		}
		return -1, false // real code before any header comment
	}
	return -1, false
}

// AssertReasonPredicates fails on raw kratos error-code checks — the
// generated reason predicate is the only legal spelling.
func AssertReasonPredicates(t *testing.T, files map[string]string) {
	t.Helper()
	for f, src := range files {
		if hit := kratosCode.FindString(src); hit != "" {
			t.Errorf("%s kratos code check %q — use the generated reason predicate", f, hit)
		}
	}
}

// AssertSerializationViaUtils fails on direct encoding/json use in the
// walked layers — serialization goes through shared-go/utils only.
func AssertSerializationViaUtils(t *testing.T, files map[string]string) {
	t.Helper()
	for f, src := range files {
		if jsonCall.MatchString(src) {
			t.Errorf("%s direct encoding/json — serialize via shared-go/utils", f)
		}
	}
}

// RungCarries reports "" when the access-ladder rung carries its gate;
// otherwise a reason naming which half failed — a missing rung (the
// selector was renamed or removed) is distinct from a present rung lacking
// its gate.
func RungCarries(src, tier, gate string) string {
	lines := strings.Split(src, "\n")
	isCode := func(s string) bool {
		t := strings.TrimSpace(s)
		return t != "" && !strings.HasPrefix(t, "//")
	}
	for i, line := range lines {
		if !isCode(line) || !strings.Contains(line, tier) {
			continue
		}
		for j := i; j >= 0; j-- {
			if !strings.Contains(lines[j], "selector.Server(") {
				continue
			}
			for _, l := range lines[j : i+1] {
				if strings.Contains(l, gate) {
					return ""
				}
			}
			return "lacks " + gate
		}
		return "rung matched but no selector.Server( block found above it"
	}
	return "rung " + tier + " not found — selector renamed or removed?"
}

// AssertLayerCensus fails when an internal/ directory is neither walked by
// the comment law (dirs) nor carries a standing exemption — an unclassified
// layer would escape the comment law silently.
func AssertLayerCensus(t *testing.T, internalRoot string, dirs []string, exempt map[string]string) {
	t.Helper()
	entries, err := os.ReadDir(internalRoot)
	if err != nil {
		t.Fatalf("read internal/: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if _, ok := exempt[name]; ok {
			continue
		}
		if !slices.Contains(dirs, name) {
			t.Errorf("internal/%s is neither comment-law-walked nor exempt — classify it in guardDirs or guardExempt", name)
		}
	}
}

// AssertSpecBesideBiz fails on a module owning an aggregate (biz.go)
// without its executable spec (biz_test.go) beside it. Face and
// transport-fixture modules (service.go + module.go only) are structurally
// exempt — no biz.go means no aggregate, so no census list to rot.
func AssertSpecBesideBiz(t *testing.T, internalRoot string) {
	t.Helper()
	modules := filepath.Join(internalRoot, "module")
	entries, err := os.ReadDir(modules)
	if err != nil {
		t.Fatalf("read module dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(modules, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "biz.go")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "biz_test.go")); err != nil {
			t.Errorf("module/%s owns an aggregate (biz.go) but carries no executable spec (biz_test.go) - spec-first is the default, not an opt-in", e.Name())
		}
	}
}

// AssertBizClockPort fails on a biz.go that reads the wall clock bare
// (time.Now()) — a UC that cannot run under fakes is a layering bug. Every
// bare read goes through the UC's clock field, or carries an inline
// wall-clock note naming why the read is presentation-only.
func AssertBizClockPort(t *testing.T, internalRoot string) {
	t.Helper()
	fsys := os.DirFS(path.Join(internalRoot, "module"))
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("read module dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := fs.ReadFile(fsys, path.Join(e.Name(), "biz.go"))
		if err != nil {
			continue // no aggregate, no clock-port surface
		}
		for i, line := range strings.Split(string(raw), "\n") {
			if bareClockRE.MatchString(line) && !strings.Contains(line, wallClockNote) {
				t.Errorf("module/%s/biz.go:%d reads the wall clock bare — route it through the UC's clock field, or mark it // wall-clock: with the reason it is presentation-only: %s", e.Name(), i+1, strings.TrimSpace(line))
			}
		}
	}
}
