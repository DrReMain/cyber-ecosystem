package fileproxy

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewFileProxyUC,
	NewFileProxyRP,
	NewFileProxyService,
)
