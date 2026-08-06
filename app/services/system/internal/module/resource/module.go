package resource

import "github.com/google/wire"

// ProviderSet wires this domain module's use case, repo, and service.
var ProviderSet = wire.NewSet(
	NewResourceUC,
	NewResourceRP,
	NewResourceService,
)
