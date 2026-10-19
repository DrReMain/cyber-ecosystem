package dept

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewDeptUC,
	NewDeptRP,
	NewDeptService,
)
