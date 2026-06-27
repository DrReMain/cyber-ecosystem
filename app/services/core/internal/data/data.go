package data

import (
	"log/slog"

	"github.com/google/wire"

	"cyber-ecosystem/app/services/core/internal/platform"
)

type RP struct {
	log      *slog.Logger
	platform *platform.Platform
}

var ProviderSet = wire.NewSet(
	NewResourceRP,
	NewUserRP,
)
