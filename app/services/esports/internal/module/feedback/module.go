package feedback

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewFeedbackUC,
	NewFeedbackAdminUC,
	NewFeedbackRP,
	NewFeedbackService,
)
