package entitle

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewCodeUC,
	NewEntitleUC,
	NewActivationCodeRP,
	NewEntitlementRP,
	NewReportPolicyRP,
	NewEntitlementService,
)
