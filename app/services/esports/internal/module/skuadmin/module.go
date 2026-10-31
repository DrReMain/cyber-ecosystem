package skuadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewSkuAdminService,
)
