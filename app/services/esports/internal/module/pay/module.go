package pay

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrderUC,
	NewSkuUC,
	NewOrderRP,
	NewPayChannelRP,
	NewSkuRP,
	NewOrderService,
	NewSkuService,
)
