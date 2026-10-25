package pay

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrderUC,
	NewSkuUC,
	NewPayChannelRP,
	NewOrderRP,
	NewSkuRP,
	NewOrderService,
	NewSkuService,
)
