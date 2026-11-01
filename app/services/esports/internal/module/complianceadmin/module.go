package complianceadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewComplianceAdminService,
)
