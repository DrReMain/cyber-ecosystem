package guards

import (
	"os"
	"path/filepath"
	"testing"
)

// Spec-first presence guard: a module that owns an aggregate (biz.go) carries
// its executable spec beside it (biz_test.go). Face and transport-fixture
// modules (service.go + module.go only) are structurally exempt — no biz.go
// means no aggregate, so no census list to rot.
func TestBizModuleCarriesSpecBesideIt(t *testing.T) {
	modules := filepath.Join("..", "module")
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
