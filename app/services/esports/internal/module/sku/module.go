package sku

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewSkuUC,
	NewSkuRP,
	NewSkuService,
)
