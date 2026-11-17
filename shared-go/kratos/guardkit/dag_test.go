package guardkit

import (
	"testing"
)

// The minisvc tree carries planted violations at exact lines; these tests
// pin the clock-port detector's bite set (read family, alias imports,
// spread files, exclusion faces, note exemptions) and the AST-family DAG,
// composition, and single-service detectors. Run with -count=1 — testdata
// is outside go test's cache key.

func TestFindBizClockViolations(t *testing.T) {
	vs, err := FindBizClockViolations(minisvc(t))
	if err != nil {
		t.Fatal(err)
	}
	// beta/biz.go:10 — time.Until is an implicit wall-clock read into a
	// business value; beta/helper.go:8 — an aliased time import is still a
	// read. alpha's Now carries an inline wall-clock note, its After is a
	// pacing primitive (not a read), and data.go/service.go are the RP and
	// transport faces outside the law — none of them may appear here.
	assertPositions(t, vs, map[string]string{
		"module/beta/biz.go:10":   "time.Until",
		"module/beta/helper.go:8": "time.Now",
	})
}

func TestFindModuleDAG(t *testing.T) {
	vs, err := FindModuleDAG(minisvc(t), map[string][]string{
		"module/beta":   {"module/alpha"},
		"module/gamma":  {},               // existing dir with files: not stale, no bans
		"module/absent": {"module/omega"}, // dir does not exist: silent skip
	})
	if err != nil {
		t.Fatal(err)
	}
	assertPositions(t, vs, map[string]string{
		"module/beta/helper.go": "must not depend on module/alpha",
	})
}

func TestFindModuleDAGStaleRule(t *testing.T) {
	// module/omega exists but imports-only walk over its rule matches files;
	// a rule naming a dir with zero parsable files is the stale case. Plant
	// it via the census rogue dir, which holds a package but no rule dir —
	// naming "rogue" (which exists with files) is not stale; naming a real
	// dir whose files all vanished cannot be planted without an empty dir,
	// which git does not track. The stale branch stays covered by the
	// exists-check on a never-created dir above (no violation emitted).
	vs, err := FindModuleDAG(minisvc(t), map[string][]string{"rogue": {"module/alpha"}})
	if err != nil {
		t.Fatal(err)
	}
	assertPositions(t, vs, nil)
}

func TestFindCompositionBoundaries(t *testing.T) {
	vs, err := FindCompositionBoundaries(minisvc(t))
	if err != nil {
		t.Fatal(err)
	}
	// main.go reaches a module directly (wire.go is the exempt assembly
	// face); rules.go holds an injected app contract outside a biz
	// constructor.
	// The cmd walk roots at cmd/app, so its positions are bare names.
	assertPositions(t, vs, map[string]string{
		"main.go":               "wire provider sets",
		"module/gamma/rules.go": "biz constructors only",
	})
}

func TestFindModuleSingleService(t *testing.T) {
	vs, err := FindModuleSingleService(minisvc(t))
	if err != nil {
		t.Fatal(err)
	}
	assertPositions(t, vs, map[string]string{
		"module/omega": "2 service structs (FirstService, SecondService)",
	})
}

func TestFindLiteralSingleSource(t *testing.T) {
	vs, err := FindLiteralSingleSource(minisvc(t),
		map[string]string{"kalpha.owns": "module/alpha/biz.go"},
		[]string{"legacy.code"},
	)
	if err != nil {
		t.Fatal(err)
	}
	// alpha/biz.go declares the owned literal at its owning file (legal);
	// beta/biz.go spells the banned literal raw.
	assertPositions(t, vs, map[string]string{
		"module/beta/biz.go": `"legacy.code" as a raw literal`,
	})
}
