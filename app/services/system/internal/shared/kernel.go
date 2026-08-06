package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/system/internal/platform"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

// ---------------------------------------------------------------------------------------------------------------------

type UC struct {
	Log *slog.Logger
	Tm  Transaction
}

func NewUC(log *slog.Logger, tm Transaction) UC {
	return UC{Log: log, Tm: tm}
}

type RP struct {
	Log      *slog.Logger
	Platform *platform.Platform
}

func NewRP(log *slog.Logger, platform *platform.Platform) RP {
	return RP{Log: log, Platform: platform}
}

// ---------------------------------------------------------------------------------------------------------------------

const DefaultTenant = "default"
