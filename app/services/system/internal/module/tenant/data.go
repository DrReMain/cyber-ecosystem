package tenant

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	entrole "cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/ent/tenant"
	entuser "cyber-ecosystem/app/services/system/internal/ent/user"
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

func (rp *tenantRP) Create(ctx context.Context, in *CreateIn) (string, error) {
	// The code's uniqueness is the database's own law — riding the
	// constraint beats a pre-read, which the caller's tenant scope would
	// blind anyway.
	client := rp.Platform.GetClient(ctx)
	t, err := client.Tenant.Create().SetCode(in.Code).SetName(in.Name).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return "", systempb.ErrorSystemTenantCodeTaken("").WithCause(err)
		}
		return "", rp.Platform.HandleEntError(err)
	}
	roleR, err := client.Role.Create().
		SetTenantID(in.Code).
		SetCode(shared.SuperAdminRoleCode).
		SetName("Super Admin").
		Save(ctx)
	if err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	if _, err = client.Permission.Create().
		SetTenantID(in.Code).
		SetRoleID(roleR.ID).
		SetOperation(shared.SuperAdminPattern).
		SetScopeKind(shared.SuperAdminScope).
		Save(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return t.ID, nil
}

func (rp *tenantRP) AddSuperAdmin(ctx context.Context, in *AdminIn) (string, error) {
	// The platform admin's own subject scopes reads to their home tenant;
	// this pass is over the target tenant's domain — the explicit switch
	// outranks the subject on every axis (reads, update/delete predicates,
	// the back-fill hook), so the whole write lands in the tenant's rows.
	ctx = security.WithTenant(ctx, in.TenantCode)
	client := rp.Platform.GetClient(ctx)
	roleR, err := client.Role.Query().
		Where(entrole.TenantID(in.TenantCode), entrole.Code(shared.SuperAdminRoleCode)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", systempb.ErrorSystemTenantNotFound("").WithCause(
				fmt.Errorf("tenant %q was never opened", in.TenantCode))
		}
		return "", rp.Platform.HandleEntError(err)
	}
	// An existing account is seated and bound as-is: identity is global,
	// passwords belong to admin flows, the seat and binding are this
	// tenant's own grants.
	u, err := client.User.Query().Where(entuser.AccountEQ(in.Account)).Only(ctx)
	if ent.IsNotFound(err) {
		var hash string
		if hash, err = utils.Hash(in.Password); err != nil {
			return "", rp.Platform.HandleEntError(err)
		}
		if u, err = client.User.Create().
			SetAccount(in.Account).
			SetPasswordHash(hash).
			SetEnabled(true).
			Save(ctx); err != nil {
			if !ent.IsConstraintError(err) {
				return "", rp.Platform.HandleEntError(err)
			}
			// A concurrent writer seated the account first (another admin,
			// a double submit): the unique broke the tie — join their user.
			if u, err = client.User.Query().Where(entuser.AccountEQ(in.Account)).Only(ctx); err != nil {
				return "", rp.Platform.HandleEntError(err)
			}
		}
	} else if err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	if _, err = client.Membership.Create().
		SetUserID(u.ID).
		SetTenantID(in.TenantCode).
		SetPopulation(security.DefaultPopulation).
		Save(ctx); err != nil && !ent.IsConstraintError(err) {
		return "", rp.Platform.HandleEntError(err)
	}
	if _, err = client.PrincipalRole.Create().
		SetTenantID(in.TenantCode).
		SetPrincipalType(shared.PrincipalTypeUser).
		SetPrincipalID(u.ID).
		SetRoleID(roleR.ID).
		Save(ctx); err != nil && !ent.IsConstraintError(err) {
		return "", rp.Platform.HandleEntError(err)
	}
	// A new binding reshapes the global authz snapshot: the engines
	// recompile on the next decision after this broadcast.
	shared.NotifyPolicyChanged(ctx, rp.Platform.GetCache(), rp.Log)
	return u.ID, nil
}

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

func (rp *tenantRP) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	query := rp.Platform.GetClient(ctx).Tenant.Query()
	helper.Where(query, in.Code != nil, func() predicate.Tenant { return tenant.CodeContainsFold(*in.Code) })
	helper.Where(query, in.Name != nil, func() predicate.Tenant { return tenant.NameContainsFold(*in.Name) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"code":      func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldCode)) },
		"name":      func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(tenant.FieldUpdatedAt)) },
	})
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListOut{
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
