package file

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewFileUC,
	NewFileRP,
	NewFileService,
)
