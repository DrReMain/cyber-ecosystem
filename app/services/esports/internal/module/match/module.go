package match

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMatchUC,
	NewMatchRP,
	NewSessionLaw,
	NewMatchService,
)
