package pay

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrderUC,
	NewSkuUC,
	NewPayChannelRP,
	NewPayOrderRP,
	NewSkuRP,
	NewOrderService,
	NewSkuService,
)
