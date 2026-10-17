package tenant

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/tenant"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type tenantRP struct {
	shared.RP
}

func NewTenantRP(logger *slog.Logger, p *platform.Platform) TenantRP {
	return &tenantRP{
		RP: shared.NewRP(logger.With("module", "module/tenant_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *tenantRP) UpdateStatus(ctx context.Context, id string, enabled bool) (*Tenant, error) {
	r, err := rp.Platform.GetClient(ctx).Tenant.UpdateOneID(id).SetEnabled(enabled).Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapTenant(r), nil
}

func (rp *tenantRP) FindByCode(ctx context.Context, code string) (*Tenant, error) {
	r, err := rp.Platform.GetClient(ctx).Tenant.Query().Where(tenant.Code(code)).Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapTenant(r), nil
}

func (rp *tenantRP) FindByID(ctx context.Context, id string) (*Tenant, error) {
	r, err := rp.Platform.GetClient(ctx).Tenant.Query().Where(tenant.ID(id)).Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapTenant(r), nil
}

func (rp *tenantRP) ListTenantCodes(ctx context.Context) ([]string, error) {
	codes, err := rp.Platform.GetClient(ctx).Tenant.Query().
		Select(tenant.FieldCode).
		Strings(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return codes, nil
}

func (rp *tenantRP) List(ctx context.Context, in *TenantListIn) (*TenantListOut, error) {
	query := rp.Platform.GetClient(ctx).Tenant.Query()
	helper.Where(query, in.Code != nil, func() predicate.Tenant { return tenant.CodeContainsFold(*in.Code) })
	helper.Where(query, in.Name != nil, func() predicate.Tenant { return tenant.NameContainsFold(*in.Name) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"code":      func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldCode)) },
		"name":      func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldUpdatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &TenantListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(rs, mapTenant),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapTenant(d *ent.Tenant) *Tenant {
	return &Tenant{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Code:      d.Code,
		Name:      d.Name,
		Enabled:   d.Enabled,
	}
}
