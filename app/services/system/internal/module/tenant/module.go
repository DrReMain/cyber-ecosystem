package tenant

import (
	"github.com/google/wire"

	"cyber-ecosystem/app/services/system/internal/shared"
)

var ProviderSet = wire.NewSet(
	NewTenantUC,
	NewTenantRP,
	NewTenantService,
	wire.Bind(new(shared.TenantDomainSource), new(*TenantUC)),
)
