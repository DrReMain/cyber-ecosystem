package game

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewGameRP,
	NewGameUC,
	NewGameService,
)
