package shared

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/system/internal/platform"
)

const PrincipalTypeUser = "user"

// The super-administrator provisioning vocabulary, shared by the boot seed
// and the runtime tenant-opening faces — one copy so every writer shapes
// the same role (no superuser code path exists; the pattern is a grant like
// any other).
const (
	SuperAdminRoleCode = "superadmin"
	SuperAdminPattern  = "/*"
	SuperAdminScope    = "all"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

type Clock func() time.Time

type HookRegistry interface {
	OnStart(fn func(context.Context) error)
	OnStop(fn func(context.Context) error)
}

type TenantDomainSource interface {
	ListTenantCodes(ctx context.Context) ([]string, error)
}

type PasswordPolicy interface {
	PasswordPolicyError(candidate string) error
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
