package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/agent/internal/platform"
)

const ExtraAccount = "account"

// MinMasterKeyBytes is the sealing master-key floor shared by every
// consumer: HKDF derivation cannot add entropy, so a shorter key would
// silently weaken everything sealed under it — construction refuses it.
const MinMasterKeyBytes = 16

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

type HookRegistry interface {
	OnStart(fn func(context.Context) error)
	OnStop(fn func(context.Context) error)
}

// Base ----------------------------------------------------------------------------------------------------------------

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
