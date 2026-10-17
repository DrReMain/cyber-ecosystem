package tenant

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Tenant struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Code      string
	Name      string
	Enabled   bool
}

type TenantListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Code    *string
	Name    *string
}

type TenantListOut struct {
	*commonpb.PageResponse
	List []*Tenant
}

// Port ----------------------------------------------------------------------------------------------------------------

type TenantRP interface {
	UpdateStatus(ctx context.Context, id string, enabled bool) (*Tenant, error)
	List(ctx context.Context, in *TenantListIn) (*TenantListOut, error)
	ListTenantCodes(ctx context.Context) ([]string, error)
	FindByCode(ctx context.Context, code string) (*Tenant, error)
	FindByID(ctx context.Context, id string) (*Tenant, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type TenantUC struct {
	shared.UC
	tenantRP TenantRP
}

func NewTenantUC(logger *slog.Logger, tm shared.Transaction, tenantRP TenantRP) *TenantUC {
	return &TenantUC{
		UC:       shared.NewUC(logger.With("module", "module/tenant"), tm),
		tenantRP: tenantRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *TenantUC) UpdateStatus(ctx context.Context, id string, enabled bool) (*Tenant, error) {
	t, err := uc.ensureTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	// The default tenant is the degenerate mode's whole world — disabling it
	// locks out every inhabitant including the admin who could undo it.
	if t.Code == security.DefaultTenant && !enabled {
		return nil, systempb.ErrorSystemTenantProtected("")
	}
	if t.Enabled == enabled {
		return t, nil
	}
	return uc.tenantRP.UpdateStatus(ctx, id, enabled)
}

func (uc *TenantUC) List(ctx context.Context, in *TenantListIn) (*TenantListOut, error) {
	return uc.tenantRP.List(ctx, in)
}

func (uc *TenantUC) ListTenantCodes(ctx context.Context) ([]string, error) {
	return uc.tenantRP.ListTenantCodes(ctx)
}

func (uc *TenantUC) Get(ctx context.Context, id string) (*Tenant, error) {
	return uc.ensureTenant(ctx, id)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *TenantUC) ensureTenant(ctx context.Context, id string) (*Tenant, error) {
	t, err := uc.tenantRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemTenantNotFound("")
		}
		return nil, err
	}
	return t, nil
}
