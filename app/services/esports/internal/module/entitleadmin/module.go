package entitleadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMembershipUC,
	NewSkuFactsRP,
	NewEntitleAdminService,
)
