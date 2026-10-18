package bootstrap

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewLifecycle,
)
