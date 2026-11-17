package guardkit

import (
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The testdata trees below carry planted violations; each test asserts the
// exact violation set so a detector that silently stops biting fails here.
// Note: testdata files are not Go sources and do not enter go test's cache
// key — when iterating on samples, run this package with -count=1 (the
// standing guard-chain discipline) or a stale PASS will lie to you.

// assertPositions fails unless the violations landed on exactly the wanted
// positions (sorted — the Find detectors' order) with each message carrying
// its wanted substring. The want side is keyed by position: map lookups
// instead of cross-slice indexing, which is both the check and the shape
// gosec can prove safe.
func assertPositions(t *testing.T, got []Violation, want map[string]string) {
	t.Helper()
	gotPos := make([]string, len(got))
	for i, v := range got {
		gotPos[i] = v.Pos
	}
	wantPos := slices.Sorted(maps.Keys(want))
	if !slices.Equal(gotPos, wantPos) {
		t.Fatalf("violation positions %v, want %v: %s", gotPos, wantPos, joinMessages(got))
	}
	for _, v := range got {
		if !strings.Contains(v.Message, want[v.Pos]) {
			t.Fatalf("violation on %s: message %q does not contain %q", v.Pos, v.Message, want[v.Pos])
		}
	}
}

func joinMessages(vs []Violation) string {
	var b strings.Builder
	for i, v := range vs {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(v.Message)
	}
	return b.String()
}

func minisvc(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "minisvc", "internal")
}

func TestFindNoDeclarationComments(t *testing.T) {
	files := map[string]string{
		"header.go": "package p\n\n// a floating prose block\nimport \"fmt\"\n",
		// The decl comment sits after the import block: a comment directly
		// above the package's first non-import statement is header-position
		// too, and a headerless file would double-report — the import block
		// isolates the declaration-comment shape.
		"decl.go":  "package p\n\nimport \"fmt\"\n\n// sits on a declaration\nfunc f() {}\n",
		"legal.go": "package p\n\nimport \"fmt\"\n\n// divider -----------------------------------------------------------------\n\nfunc f() {\n\t// in-body why\n\tfmt.Println(1)\n}\n",
	}
	assertPositions(t, FindNoDeclarationComments(files), map[string]string{
		"decl.go:5":   "declaration comment",
		"header.go:3": "file-header comment",
	})
}

func TestFindReasonPredicates(t *testing.T) {
	assertPositions(t, FindReasonPredicates(map[string]string{
		"bad.go": "if errors.IsNotFound(err) {\n\treturn err\n}\n",
		"ok.go":  "if errorspb.IsUserNotFound(err) {\n\treturn err\n}\n",
	}), map[string]string{"bad.go": "errors.IsNotFound"})
}

func TestFindSerializationViaUtils(t *testing.T) {
	assertPositions(t, FindSerializationViaUtils(map[string]string{
		"bad.go": "b, err := json.Marshal(x)\n",
		"ok.go":  "b, err := utils.Marshal(x)\n",
	}), map[string]string{"bad.go": "encoding/json"})
}

func TestFindLayerCensus(t *testing.T) {
	vs, err := FindLayerCensus(minisvc(t), []string{"module"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertPositions(t, vs, map[string]string{"rogue": "neither comment-law-walked nor exempt"})

	vs, err = FindLayerCensus(minisvc(t), []string{"module"}, map[string]string{"rogue": "sample exemption"})
	if err != nil {
		t.Fatal(err)
	}
	assertPositions(t, vs, nil)
}

func TestFindSpecBesideBiz(t *testing.T) {
	vs, err := FindSpecBesideBiz(minisvc(t))
	if err != nil {
		t.Fatal(err)
	}
	// alpha carries its spec; gamma and omega own no aggregate (no biz.go).
	assertPositions(t, vs, map[string]string{"module/beta": "no executable spec"})
}

func TestRungCarries(t *testing.T) {
	src := "m := selector.Server(\n\tkrauth.SessionAuth(), // BUILTIN\n)\n"
	if got := RungCarries(src, "BUILTIN", "krauth.SessionAuth"); got != "" {
		t.Fatalf("gate carried, got reason %q", got)
	}
	if got := RungCarries(src, "BUILTIN", "security.MissingGate"); got != "lacks security.MissingGate" {
		t.Fatalf("missing gate: %q", got)
	}
	if got := RungCarries(src, "ADMIN", "krauth.SessionAuth"); got != "rung ADMIN not found — selector renamed or removed?" {
		t.Fatalf("missing rung: %q", got)
	}
}

func TestSortedDeterminism(t *testing.T) {
	files := map[string]string{
		"z.go": "b, err := json.Marshal(x)\n",
		"a.go": "b, err := json.Marshal(x)\n",
		"m.go": "b, err := json.Marshal(x)\n",
	}
	for range 3 {
		vs := FindSerializationViaUtils(files)
		var pos []string
		for _, v := range vs {
			pos = append(pos, v.Pos)
		}
		if !slices.Equal(pos, []string{"a.go", "m.go", "z.go"}) {
			t.Fatalf("order not deterministic: %v", pos)
		}
	}
}
