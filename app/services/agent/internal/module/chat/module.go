package chat

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewChatUC,
	NewChatRP,
	NewChatService,
)
