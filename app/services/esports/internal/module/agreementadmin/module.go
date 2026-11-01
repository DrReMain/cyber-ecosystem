package agreementadmin

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAgreementAdminService,
)
