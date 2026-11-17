package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/guardkit"
)

// Spec-first presence guard: a module that owns an aggregate (biz.go)
// carries its executable spec beside it (biz_test.go) — the walk and the
// exemption rationale live in guardkit.AssertSpecBesideBiz.
func TestBizModuleCarriesSpecBesideIt(t *testing.T) {
	guardkit.AssertSpecBesideBiz(t, "..")
}
