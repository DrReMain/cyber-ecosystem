package profile

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewProfileUC,
	NewProfileRP,
	NewProfileService,
)
