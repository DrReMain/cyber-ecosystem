package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/template/services/base/internal/platform"
)

type RP struct {
	log      *log.Helper
	platform *platform.Platform
}

var ProviderSet = wire.NewSet(
	NewResourceRP,
	NewMessageRP,
)
