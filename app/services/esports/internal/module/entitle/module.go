package entitle

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewCodeUC,
	NewEntitleUC,
	NewCodeRP,
	NewEntitlementRP,
	NewPolicyRP,
	NewActivationCodeService,
	NewEntitlementService,
	// The UC's completion seam, bound here at the producer: consumers see
	// the narrow hook, the provider graph stays a plain tree.
	wire.Bind(new(TrialGrantHook), new(*EntitleUC)),
)
