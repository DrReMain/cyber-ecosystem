package test

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewTestUC,
	NewTestRP,
	NewTestService,
)
