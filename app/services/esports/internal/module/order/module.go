package order

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrderUC,
	NewOrderRP,
	NewPayChannelRP,
	NewOrderService,
)
