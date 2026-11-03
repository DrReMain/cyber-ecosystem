package entitle

import (
	"github.com/google/wire"

	"cyber-ecosystem/app/services/esports/internal/module/profile"
)

var ProviderSet = wire.NewSet(
	NewCodeUC,
	NewEntitleUC,
	NewActivationCodeRP,
	NewEntitlementRP,
	NewReportPolicyRP,
	NewEntitlementService,
	wire.Bind(new(profile.MembershipRP), new(*EntitleUC)),
)
