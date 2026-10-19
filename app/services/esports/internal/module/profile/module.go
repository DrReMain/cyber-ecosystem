package profile

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewProfileRP,
	NewProfileUC,
	NewProfileService,
)
