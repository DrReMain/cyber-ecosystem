package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/platform"
)

type RP struct {
	log      *log.Helper
	platform *platform.Platform
}

var ProviderSet = wire.NewSet(
	NewArticleRP,
	NewResourceRP,
)
