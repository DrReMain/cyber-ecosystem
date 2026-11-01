package agreement

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAgreementUC,
	NewAgreementAdminUC,
	NewAgreementRP,
	NewAgreementService,
)
