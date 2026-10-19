package agentconfig

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAgentConfigUC,
	NewAgentConfigRP,
	NewAgentConfigService,
)
