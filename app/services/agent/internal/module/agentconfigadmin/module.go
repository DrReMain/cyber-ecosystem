package agentconfigadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAgentConfigAdminService,
)
