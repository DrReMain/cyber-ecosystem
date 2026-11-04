package message

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMessageUC,
	NewMessageAdminUC,
	NewMessageRP,
	NewMessageService,
)
