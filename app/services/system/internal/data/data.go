package data

import (
	"log/slog"

	"github.com/google/wire"

	"cyber-ecosystem/app/services/system/internal/platform"
)

type RP struct {
	log      *slog.Logger
	platform *platform.Platform
}

var ProviderSet = wire.NewSet(
	NewItemRP,
	NewResourceRP,
	NewTokenRP,
	NewUserRP,
)
