package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/app_1/services/service_2/internal/biz"
)

// RP is the base struct for data layer repositories.
type RP struct {
	log   *log.Helper
	store *Store
}

var ProviderSet = wire.NewSet(
	NewStore,
	NewCache,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Store)),
	// -----------------------------------------------------------------------------------------------------------------
	NewGRPCClientService1,
	NewConnectClientService1,
	NewReadingRP,
)
