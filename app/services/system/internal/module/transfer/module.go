package transfer

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewTransferService,
)
