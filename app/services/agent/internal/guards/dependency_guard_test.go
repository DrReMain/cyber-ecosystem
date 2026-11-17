package guards

import (
	"testing"

	"cyber-ecosystem/shared-go/kratos/guardkit"
)

// Static guards for the dependency rules as code: the module DAG and the
// composition layering (cmd/app wires through bootstrap; shared stays a
// contracts-only kernel; injected app contracts land in module biz
// constructors). The walk machinery lives in shared-go/kratos/guardkit.

// bannedImports maps a package directory (relative to internal/, prefix
// match) to import markers it must not contain. Direct imports only: a
// transitive leak requires importing the middleman, which these rules catch
// at that hop.
var bannedImports = map[string][]string{
	// The engine floor is module-free infrastructure and never reads
	// business aggregates.
	"module/authz": {"module/agentconfig", "module/agentconfigadmin", "module/chat"},
	// The operator face is a strippable unit; the business module must
	// survive its deletion untouched.
	"module/agentconfig": {"module/agentconfigadmin"},
	// ACL adapters implement biz-declared ports (importing the declaring
	// module for the interface is the designed direction; the reverse is the
	// leak), and stay ent-free.
	"client": {"internal/ent"},
	// The kernel-to-schema direction is one-way (shared → platform → ent);
	// an import back closes the cycle.
	"ent/schema": {"internal/shared"},
	// Bootstrap sits below business modules.
	"bootstrap": {"module/"},
	// The kernel stays ent-free.
	"shared": {"internal/ent"},
}

func TestModuleDependencyDAG(t *testing.T) {
	guardkit.AssertModuleDAG(t, "..", bannedImports)
}

func TestCompositionBoundaries(t *testing.T) {
	guardkit.AssertCompositionBoundaries(t, "..")
}

func TestModuleSingleService(t *testing.T) {
	guardkit.AssertModuleSingleService(t, "..")
}
