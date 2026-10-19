package policy

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/authzpolicy"
	"cyber-ecosystem/app/services/system/internal/ent/permissionpolicy"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type policyRP struct {
	shared.RP
}

func NewPolicyRP(logger *slog.Logger, p *platform.Platform) PolicyRP {
	return &policyRP{
		RP: shared.NewRP(logger.With("module", "module/policy_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *policyRP) Create(ctx context.Context, p *Policy) (*Policy, error) {
	created, err := rp.Platform.GetClient(ctx).AuthzPolicy.Create().
		SetKind(p.Kind).
		SetName(p.Name).
		SetParams(p.Params).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapPolicy(created), nil
}

func (rp *policyRP) Update(ctx context.Context, fieldsMask []string, p *Policy) (*Policy, error) {
	updater := rp.Platform.GetClient(ctx).AuthzPolicy.UpdateOneID(p.ID)
	helper.Handler{
		"name": {
			Condition: p.Name != "",
			OnTrue:    func() { updater.SetName(p.Name) },
			OnFalse:   func() {},
		},
		"enabled": {
			Condition: true,
			OnTrue:    func() { updater.SetEnabled(p.Enabled) },
			OnFalse:   func() {},
		},
		"params": {
			Condition: p.Params != nil,
			OnTrue:    func() { updater.SetParams(p.Params) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)
	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapPolicy(updated), nil
}

func (rp *policyRP) Delete(ctx context.Context, id string) (string, error) {
	client := rp.Platform.GetClient(ctx)
	// Grants keep their other policies; the join rows go with the deleted
	// policy. Same client so both ride the caller's transaction.
	if _, err := client.PermissionPolicy.Delete().Where(permissionpolicy.PolicyID(id)).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	if err := client.AuthzPolicy.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *policyRP) List(ctx context.Context, in *PolicyListIn) (*PolicyListOut, error) {
	query := rp.Platform.GetClient(ctx).AuthzPolicy.Query()
	helper.Where(query, in.Kind != nil, func() predicate.AuthzPolicy { return authzpolicy.KindEQ(*in.Kind) })
	helper.Where(query, in.Name != nil, func() predicate.AuthzPolicy { return authzpolicy.NameContainsFold(*in.Name) })
	helper.Where(query, in.Enabled != nil, func() predicate.AuthzPolicy { return authzpolicy.EnabledEQ(*in.Enabled) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"kind":      func(sel helper.SQLSelector) { query.Order(sel(authzpolicy.FieldKind)) },
		"name":      func(sel helper.SQLSelector) { query.Order(sel(authzpolicy.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(authzpolicy.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(authzpolicy.FieldUpdatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	ps, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := utils.SliceMap(ps, mapPolicy)
	if err := attachBoundCounts(ctx, rp.Platform.GetClient(ctx), out); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &PolicyListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         out,
	}, nil
}

func (rp *policyRP) FindByID(ctx context.Context, id string) (*Policy, error) {
	client := rp.Platform.GetClient(ctx)
	d, err := client.AuthzPolicy.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	p := mapPolicy(d)
	if err := attachBoundCounts(ctx, client, []*Policy{p}); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return p, nil
}

func (rp *policyRP) FindByIDs(ctx context.Context, ids []string) ([]*Policy, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Tenant scoping comes from the mixin interceptor, so ids outside the
	// caller's tenant come back absent — callers report them as not found.
	ps, err := rp.Platform.GetClient(ctx).AuthzPolicy.Query().Where(authzpolicy.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(ps, mapPolicy), nil
}

func (rp *policyRP) NotifyChanged(ctx context.Context) {
	shared.NotifyPolicyChanged(ctx, rp.Platform.GetCache(), rp.Log)
}

// Private -------------------------------------------------------------------------------------------------------------

func mapPolicy(d *ent.AuthzPolicy) *Policy {
	return &Policy{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		Kind:      d.Kind,
		Name:      d.Name,
		Params:    d.Params,
		Enabled:   d.Enabled,
	}
}

func attachBoundCounts(ctx context.Context, client *ent.Client, ps []*Policy) error {
	if len(ps) == 0 {
		return nil
	}
	ids := make([]string, 0, len(ps))
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	var rows []struct {
		PolicyID string `json:"policy_id"`
		Count    int    `json:"count"`
	}
	if err := client.PermissionPolicy.Query().Where(permissionpolicy.PolicyIDIn(ids...)).
		GroupBy(permissionpolicy.FieldPolicyID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return err
	}
	counts := make(map[string]int64, len(rows))
	for _, r := range rows {
		counts[r.PolicyID] = int64(r.Count)
	}
	for _, p := range ps {
		p.BoundCount = counts[p.ID]
	}
	return nil
}
