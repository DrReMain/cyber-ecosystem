package order

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrderUC,
	NewOrderAdminUC,
	NewOrderRP,
	NewPayChannelRP,
	NewOrderService,
	NewMockCashier,
)
