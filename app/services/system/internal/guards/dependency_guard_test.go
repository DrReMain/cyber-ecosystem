package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/guardkit"
)

// Static guards for the dependency rules as code: the module DAG, the
// shared-contract seams, and the composition layering (cmd/app wires
// through bootstrap; shared stays a contracts-only kernel; injected app
// contracts land in module biz constructors). A pure-function property like
// compile() is review-enforced — a file walk cannot see it; the rest is
// checked here, plus single-sourcing of the ABAC vocabulary. The walk
// machinery lives in shared-go/kratos/guardkit.

// bannedImports maps a package directory (relative to internal/, prefix
// match) to import markers it must not contain. Direct imports only: a
// transitive leak requires importing the middleman, which these rules catch
// at that hop.
var bannedImports = map[string][]string{
	// The engine reads ent tables directly and never sibling modules.
	// Sibling census: this list must name every module/* directory — a new
	// module missing here loses the ban silently (the enumeration-rot the
	// mixin census also guards against).
	"module/authz": {
		"module/role", "module/policy", "module/dept", "module/user",
		"module/auth", "module/resource", "module/transfer",
		"module/tenant", "module/population",
	},
	// role may read the authz port (diagnostics trio); the others must not.
	"module/policy": {"module/authz"},
	"module/dept":   {"module/authz"},
	"module/user":   {"module/authz", "module/auth"},
	"module/role":   {"module/user"},
	// Seam 1: the session-auth closure stays free of ent.
	"module/auth": {"internal/ent"},
	// The kernel-to-schema direction is one-way (shared → platform → ent);
	// an import back closes the cycle.
	"ent/schema": {"internal/shared"},
	// Bootstrap sits below business modules.
	"bootstrap": {"module/"},
	// The kernel stays ent-free (Paginate is generic by design).
	"shared": {"internal/ent"},
}

// literalOwners pins contract strings to the single file allowed to declare
// them; bannedLiterals must not appear in service code at all — the kauthz
// constant is the only spelling.
var (
	literalOwners = map[string]string{
		"authz:policy:ver":     "shared/policy_notify.go",
		"authz:policy:changed": "shared/policy_notify.go",
		"auth:web:revoked:":    "shared/session_revoke.go",
	}
	bannedLiterals = []string{"time_window", "calendar", "time.now"}
)

func TestModuleDependencyDAG(t *testing.T) {
	guardkit.AssertModuleDAG(t, "..", bannedImports)
}

func TestCompositionBoundaries(t *testing.T) {
	guardkit.AssertCompositionBoundaries(t, "..")
}

func TestSharedContractSingleSource(t *testing.T) {
	guardkit.AssertLiteralSingleSource(t, "..", literalOwners, bannedLiterals)
}

func TestModuleSingleService(t *testing.T) {
	guardkit.AssertModuleSingleService(t, "..")
}
