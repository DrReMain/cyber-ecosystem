package filepresign

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewFilePresignUC,
	NewFilePresignRP,
	NewFilePresignService,
)
