package role

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewRoleUC,
	NewRoleRP,
	NewRoleService,
)
