package shared

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/esports/internal/platform"
)

const ExtraAccount = "account"

// The entitlement ledger's scope vocabulary — one home because two
// modules write it (pay grants on settle, entitle grants on redeem) and a
// drifted literal would silently split the ledger.
const (
	ScopeTrial      = "trial"
	ScopeMembership = "membership"
	ScopeReportType = "report_type"
)

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
