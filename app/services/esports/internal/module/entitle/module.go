package entitle

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewCodeUC,
	NewEntitleUC,
	NewCodeRP,
	NewEntitlementRP,
	NewActivationCodeService,
	NewEntitlementService,
)
