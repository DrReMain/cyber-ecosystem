package compliance

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewComplianceUC,
	NewComplianceRP,
)
