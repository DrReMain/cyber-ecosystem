package messageadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMessageAdminService,
)
