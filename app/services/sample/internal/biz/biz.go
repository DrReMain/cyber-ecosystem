package biz

import (
	"log/slog"

	"github.com/google/wire"
)

type UC struct {
	log *slog.Logger
}

var ProviderSet = wire.NewSet(
	NewSampleUC,
)
