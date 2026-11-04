package mockcashier

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMockCashierService,
)
