package biz

import (
	"log/slog"

	"github.com/google/wire"

	"cyber-ecosystem/app/services/edge_mobile/internal/platform"
)

type UC struct {
	log *slog.Logger
	tm  platform.Transaction
}

var ProviderSet = wire.NewSet(NewResourceUC)
