package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/guardkit"
)

// Convention guards: the recurring violation classes (declaration
// comments, reason predicates, serialization, and plane gates) as machine
// gates — rows K-CMT-1, K-ERR-2, K-UTL-1, K-LAD-1
// run green here. The machinery lives in shared-go/kratos/guardkit; this
// file holds system's topology tables and the thin tests driving them.

// guardDirs are the comment-law zones; ent/schema (own comment class) and
// this kernel package keep their own conventions.
var guardDirs = []string{"module", "client", "server", "platform", "bootstrap"}

// guardExempt names the internal/ layers outside the comment-law walk, each
// with its standing reason; TestLayerCensus fails when a NEW layer appears
// that is neither walked nor exempt — an unclassified layer would escape
// the comment law silently.
var guardExempt = map[string]string{
	"conf":   "generated from conf.proto",
	"e2e":    "test-only package, never imported by production",
	"ent":    "generated + schema (own comment class)",
	"guards": "this kernel package (own conventions)",
	"shared": "kernel package (own conventions)",
}

// rungGates locks the access ladder: each rung and the gate it must carry.
var rungGates = map[string]string{
	"ACCESS_ADMIN":    "security.BackendPlaneGuard()",
	"ACCESS_PLATFORM": "security.BackendPlaneGuard()",
	"ACCESS_BUILTIN":  "krauth.SessionAuth(",
}

func layerFiles(t *testing.T) map[string]string {
	t.Helper()
	return guardkit.LayerFiles(t, "..", guardDirs)
}

func TestNoDeclarationComments(t *testing.T) {
	guardkit.AssertNoDeclarationComments(t, layerFiles(t))
}

func TestLayerCensus(t *testing.T) {
	guardkit.AssertLayerCensus(t, "..", guardDirs, guardExempt)
}

func TestReasonPredicates(t *testing.T) {
	guardkit.AssertReasonPredicates(t, layerFiles(t))
}

func TestSerializationViaUtils(t *testing.T) {
	guardkit.AssertSerializationViaUtils(t, layerFiles(t))
}

// TestAccessLadderGates locks the plane-gate law on every rung (K-LAD-1).
func TestAccessLadderGates(t *testing.T) {
	src, ok := layerFiles(t)["server.go"]
	if !ok {
		t.Fatalf("server.go missing from the guard walk")
	}
	for tier, gate := range rungGates {
		if err := guardkit.RungCarries(src, tier, gate); err != "" {
			t.Errorf("server.go %s rung: %s", tier, err)
		}
	}
}
