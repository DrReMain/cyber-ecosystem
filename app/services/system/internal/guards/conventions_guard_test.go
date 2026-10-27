package guards

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Convention guards: the recurring violation classes (declaration
// comments, reason predicates, serialization, and plane gates) as machine
// gates — rows K-CMT-1, K-ERR-2, K-UTL-1, K-LAD-1
// run green here.

var (
	guardDividerRE   = regexp.MustCompile(`^// [A-Za-z][A-Za-z0-9 /]* -{8,}\s*$`)
	guardDirectiveRE = regexp.MustCompile(`^//[a-z]+:`)
	guardDeclRE      = regexp.MustCompile(`^(type|func|const|var)\b`)
	guardKratosCode  = regexp.MustCompile(`errors\.Is(NotFound|InvalidArgument|Unauthorized|Forbidden|Conflict|BadRequest|InternalServerError|FailedPrecondition)\(`)
	guardJSON        = regexp.MustCompile(`json\.(Marshal|Unmarshal|NewEncoder|NewDecoder)\(`)
)

// guardDirs are the comment-law zones; ent/schema (own comment class) and
// this kernel package keep their own conventions.
var guardDirs = []string{"module", "client", "server", "platform", "bootstrap"}

func guardGoFiles(t *testing.T) map[string]string {
	t.Helper()
	out := make(map[string]string)
	for _, dir := range guardDirs {
		root := filepath.Join("..", dir)
		if _, statErr := os.Stat(root); statErr != nil {
			continue // a service may lack the layer (system has no client/)
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

func TestNoDeclarationComments(t *testing.T) {
	for f, src := range guardGoFiles(t) {
		lines := strings.Split(src, "\n")
		if i, ok := guardHeaderComment(lines); ok {
			t.Errorf("%s:%d file-header comment — neither legal kind lives above the imports: %q", f, i+1, strings.TrimSpace(lines[i]))
		}
		for i, line := range lines {
			s := strings.TrimSpace(line)
			if !strings.HasPrefix(s, "//") || guardDividerRE.MatchString(s) || guardDirectiveRE.MatchString(s) {
				continue
			}
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j < len(lines) && guardDeclRE.MatchString(strings.TrimSpace(lines[j])) {
				t.Errorf("%s:%d declaration comment: %q", f, i+1, s)
			}
		}
	}
}

// guardHeaderComment finds a comment line in the file-header position —
// after the package clause, before the first import or declaration. The
// comment law's two kinds are dividers and in-body why; a block floating
// above the imports is the third shape, and it escapes the declaration
// signature (the next non-blank line opens import, not a declaration).
func guardHeaderComment(lines []string) (int, bool) {
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
		if strings.HasPrefix(s, "//") && !guardDividerRE.MatchString(s) && !guardDirectiveRE.MatchString(s) {
			return i, true
		}
		return -1, false // real code before any header comment
	}
	return -1, false
}

func TestReasonPredicates(t *testing.T) {
	for f, src := range guardGoFiles(t) {
		if hit := guardKratosCode.FindString(src); hit != "" {
			t.Errorf("%s kratos code check %q — use the generated reason predicate", f, hit)
		}
	}
}

func TestSerializationViaUtils(t *testing.T) {
	for f, src := range guardGoFiles(t) {
		if guardJSON.MatchString(src) {
			t.Errorf("%s direct encoding/json — serialize via shared-go/utils", f)
		}
	}
}

// TestAdminRungPlaneGate locks the plane-gate law: the ADMIN rung seals
// to the backend plane ahead of authz (K-LAD-1).
func TestAdminRungPlaneGate(t *testing.T) {
	src, ok := guardGoFiles(t)["server.go"]
	if !ok {
		t.Fatalf("server.go missing from the guard walk")
	}
	if !rungCarries(src, "ACCESS_ADMIN", "security.BackendPlaneGuard()") {
		t.Errorf("server.go ADMIN rung lacks the backend plane gate")
	}
	if !rungCarries(src, "ACCESS_PLATFORM", "security.BackendPlaneGuard()") {
		t.Errorf("server.go PLATFORM rung lacks the backend plane gate")
	}
}

func rungCarries(src, tier, gate string) bool {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		if !strings.Contains(line, tier) {
			continue
		}
		for j := i; j >= 0; j-- {
			if !strings.Contains(lines[j], "selector.Server(") {
				continue
			}
			for _, l := range lines[j : i+1] {
				if strings.Contains(l, gate) {
					return true
				}
			}
			return false
		}
	}
	return false
}
