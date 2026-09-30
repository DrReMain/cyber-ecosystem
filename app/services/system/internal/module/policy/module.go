package policy

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewPolicyUC,
	NewPolicyRP,
	NewPolicyService,
)
