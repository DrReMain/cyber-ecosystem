package role

import (
	"context"
	"log/slog"
	"slices"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/permission"
	"cyber-ecosystem/app/services/system/internal/ent/permissionpolicy"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/principalrole"
	"cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type roleRP struct {
	shared.RP
}

func NewRoleRP(logger *slog.Logger, p *platform.Platform) RoleRP {
	return &roleRP{
		RP: shared.NewRP(logger.With("module", "module/role_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *roleRP) Create(ctx context.Context, r *Role) (*Role, error) {
	created, err := rp.Platform.GetClient(ctx).Role.Create().
		SetCode(r.Code).
		SetName(r.Name).
		SetRemark(utils.Deref(r.Remark, "")).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapRole(created), nil
}

func (rp *roleRP) Update(ctx context.Context, fieldsMask []string, r *Role) (*Role, error) {
	updater := rp.Platform.GetClient(ctx).Role.UpdateOneID(r.ID)
	helper.Handler{
		"name": {
			Condition: r.Name != "",
			OnTrue:    func() { updater.SetName(r.Name) },
			OnFalse:   func() {},
		},
		"remark": {
			Condition: r.Remark != nil,
			OnTrue:    func() { updater.SetRemark(*r.Remark) },
			OnFalse:   func() { updater.SetRemark("") },
		},
	}.Emit(fieldsMask)
	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapRole(updated), nil
}

func (rp *roleRP) UpdateStatus(ctx context.Context, id string, enabled bool) (*Role, error) {
	updated, err := rp.Platform.GetClient(ctx).Role.UpdateOneID(id).SetEnabled(enabled).Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapRole(updated), nil
}

func (rp *roleRP) Delete(ctx context.Context, id string) (string, error) {
	client := rp.Platform.GetClient(ctx)
	// Aggregate delete, all on the caller's transaction client: join rows
	// first, then the role's permissions, then the role itself.
	pids, err := client.Permission.Query().Where(permission.RoleID(id)).IDs(ctx)
	if err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	if len(pids) > 0 {
		if _, err = client.PermissionPolicy.Delete().Where(permissionpolicy.PermissionIDIn(pids...)).Exec(ctx); err != nil {
			return "", rp.Platform.HandleEntError(err)
		}
		if _, err = client.Permission.Delete().Where(permission.RoleID(id)).Exec(ctx); err != nil {
			return "", rp.Platform.HandleEntError(err)
		}
	}
	if err = client.Role.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *roleRP) List(ctx context.Context, in *RoleListIn) (*RoleListOut, error) {
	query := rp.Platform.GetClient(ctx).Role.Query()
	helper.Where(query, in.Code != nil, func() predicate.Role { return role.CodeContainsFold(*in.Code) })
	helper.Where(query, in.Name != nil, func() predicate.Role { return role.NameContainsFold(*in.Name) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"code":      func(sel helper.SQLSelector) { query.Order(sel(role.FieldCode)) },
		"name":      func(sel helper.SQLSelector) { query.Order(sel(role.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(role.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(role.FieldUpdatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := &RoleListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(rs, mapRole),
	}
	if err := rp.enrichCounts(ctx, out.List); err != nil {
		return nil, err
	}
	return out, nil
}

func (rp *roleRP) FindByID(ctx context.Context, id string) (*Role, error) {
	d, err := rp.Platform.GetClient(ctx).Role.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapRole(d), nil
}

func (rp *roleRP) ExistsBindings(ctx context.Context, roleID string) (bool, error) {
	exists, err := rp.Platform.GetClient(ctx).PrincipalRole.Query().Where(principalrole.RoleID(roleID)).Exist(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return exists, nil
}

func (rp *roleRP) CountBindings(ctx context.Context, roleID string) (int64, error) {
	n, err := rp.Platform.GetClient(ctx).PrincipalRole.Query().Where(principalrole.RoleID(roleID)).Count(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return int64(n), nil
}

func (rp *roleRP) ListByPrincipal(ctx context.Context, principalType, principalID string) ([]*Role, error) {
	client := rp.Platform.GetClient(ctx)
	bindings, err := client.PrincipalRole.Query().Where(
		principalrole.PrincipalType(principalType),
		principalrole.PrincipalID(principalID),
	).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	roleIDs := make([]string, 0, len(bindings))
	for _, b := range bindings {
		roleIDs = append(roleIDs, b.RoleID)
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}
	// Ordered by code so consumers get a stable list regardless of binding
	// insertion order.
	rows, err := client.Role.Query().Where(role.IDIn(roleIDs...)).Order(ent.Asc(role.FieldCode)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapRole), nil
}

func (rp *roleRP) ReplaceGrants(ctx context.Context, roleID string, grants []*Permission) error {
	client := rp.Platform.GetClient(ctx)
	stored, err := client.Permission.Query().Where(permission.RoleID(roleID)).All(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	storedByOp := make(map[string]*ent.Permission, len(stored))
	storedIDs := make([]string, 0, len(stored))
	for _, p := range stored {
		storedByOp[p.Operation] = p
		storedIDs = append(storedIDs, p.ID)
	}
	storedPolicies, err := rp.policyLinks(ctx, client, storedIDs)
	if err != nil {
		return err
	}
	wantedByOp := make(map[string]*Permission, len(grants))
	for _, g := range grants {
		wantedByOp[g.Operation] = g
	}
	for op, g := range wantedByOp {
		existing, ok := storedByOp[op]
		if !ok {
			created, cerr := client.Permission.Create().
				SetRoleID(roleID).
				SetOperation(op).
				SetScopeKind(g.ScopeKind).
				Save(ctx)
			if cerr != nil {
				// A racing writer may have won the unique (role_id, operation);
				// treat the constraint as convergence, not failure.
				if !ent.IsConstraintError(cerr) {
					return rp.Platform.HandleEntError(cerr)
				}
				continue
			}
			if err := rp.replacePolicyLinks(ctx, client, created.ID, g.PolicyIDs); err != nil {
				return err
			}
			continue
		}
		if existing.ScopeKind != g.ScopeKind {
			if _, err := client.Permission.UpdateOneID(existing.ID).SetScopeKind(g.ScopeKind).Save(ctx); err != nil {
				return rp.Platform.HandleEntError(err)
			}
		}
		if !samePolicySet(storedPolicies[existing.ID], g.PolicyIDs) {
			if err := rp.replacePolicyLinks(ctx, client, existing.ID, g.PolicyIDs); err != nil {
				return err
			}
		}
	}
	for op, existing := range storedByOp {
		if _, ok := wantedByOp[op]; ok {
			continue
		}
		if _, err := client.PermissionPolicy.Delete().Where(permissionpolicy.PermissionID(existing.ID)).Exec(ctx); err != nil {
			return rp.Platform.HandleEntError(err)
		}
		if err := client.Permission.DeleteOneID(existing.ID).Exec(ctx); err != nil {
			return rp.Platform.HandleEntError(err)
		}
	}
	return nil
}

func (rp *roleRP) ListGrants(ctx context.Context, roleID string) ([]*Permission, error) {
	client := rp.Platform.GetClient(ctx)
	ps, err := client.Permission.Query().Where(permission.RoleID(roleID)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := utils.SliceMap(ps, mapPermission)
	if len(ps) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(ps))
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	links, err := rp.policyLinks(ctx, client, ids)
	if err != nil {
		return nil, err
	}
	for _, g := range out {
		g.PolicyIDs = links[g.ID]
	}
	return out, nil
}

func (rp *roleRP) RemoveMember(ctx context.Context, roleID, principalType, principalID string) error {
	_, err := rp.Platform.GetClient(ctx).PrincipalRole.Delete().Where(
		principalrole.PrincipalType(principalType),
		principalrole.PrincipalID(principalID),
		principalrole.RoleID(roleID),
	).Exec(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil // deleting zero rows is success: idempotent revoke
}

func (rp *roleRP) ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error) {
	query := rp.Platform.GetClient(ctx).PrincipalRole.Query().Where(principalrole.RoleID(in.RoleID))
	helper.Where(query, in.PrincipalType != nil, func() predicate.PrincipalRole { return principalrole.PrincipalTypeEQ(*in.PrincipalType) })
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	ms, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &MemberListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(ms, mapMember),
	}, nil
}

func (rp *roleRP) NotifyChanged(ctx context.Context) {
	shared.NotifyPolicyChanged(ctx, rp.Platform.GetCache(), rp.Log)
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *roleRP) enrichCounts(ctx context.Context, roles []*Role) error {
	if len(roles) == 0 {
		return nil
	}
	client := rp.Platform.GetClient(ctx)
	ids := make([]string, len(roles))
	for i, r := range roles {
		ids[i] = r.ID
	}
	grants, err := client.Permission.Query().Where(permission.RoleIDIn(ids...)).All(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	dos := utils.SliceMap(grants, mapPermission)
	grantIDs := make([]string, 0, len(dos))
	for _, g := range dos {
		grantIDs = append(grantIDs, g.ID)
	}
	links, err := rp.policyLinks(ctx, client, grantIDs)
	if err != nil {
		return err
	}
	byRole := make(map[string][]*Permission, len(roles))
	for _, g := range dos {
		g.PolicyIDs = links[g.ID]
		byRole[g.RoleID] = append(byRole[g.RoleID], g)
	}
	for _, r := range roles {
		applyGrantCounts(r, byRole[r.ID])
	}
	var bindings []struct {
		RoleID string `json:"role_id"`
		Count  int    `json:"count"`
	}
	if err := client.PrincipalRole.Query().Where(principalrole.RoleIDIn(ids...)).
		GroupBy(principalrole.FieldRoleID).
		Aggregate(ent.Count()).
		Scan(ctx, &bindings); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	for _, b := range bindings {
		for _, r := range roles {
			if r.ID == b.RoleID {
				r.BoundCount = int64(b.Count)
				break
			}
		}
	}
	return nil
}

func (rp *roleRP) policyLinks(ctx context.Context, client *ent.Client, permissionIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(permissionIDs))
	if len(permissionIDs) == 0 {
		return out, nil
	}
	links, err := client.PermissionPolicy.Query().Where(permissionpolicy.PermissionIDIn(permissionIDs...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	for _, l := range links {
		out[l.PermissionID] = append(out[l.PermissionID], l.PolicyID)
	}
	return out, nil
}

func (rp *roleRP) replacePolicyLinks(ctx context.Context, client *ent.Client, permissionID string, policyIDs []string) error {
	// Delete-then-insert beats an intersection diff at this size — a grant
	// carries at most a handful of links.
	if _, err := client.PermissionPolicy.Delete().Where(permissionpolicy.PermissionID(permissionID)).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	for _, pid := range policyIDs {
		if _, err := client.PermissionPolicy.Create().SetPermissionID(permissionID).SetPolicyID(pid).Save(ctx); err != nil {
			// Mirror ReplaceGrants: a unique race means another writer set
			// the same link, which is the convergence we wanted anyway.
			if !ent.IsConstraintError(err) {
				return rp.Platform.HandleEntError(err)
			}
		}
	}
	return nil
}

func samePolicySet(stored, wanted []string) bool {
	if len(stored) != len(wanted) {
		return false
	}
	s := append([]string(nil), stored...)
	w := append([]string(nil), wanted...)
	slices.Sort(s)
	slices.Sort(w)
	return slices.Equal(s, w)
}

func mapRole(d *ent.Role) *Role {
	return &Role{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		Code:      d.Code,
		Name:      d.Name,
		Enabled:   d.Enabled,
		Remark:    &d.Remark,
	}
}

func mapPermission(d *ent.Permission) *Permission {
	return &Permission{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		RoleID:    d.RoleID,
		Operation: d.Operation,
		Effect:    d.Effect,
		ScopeKind: d.ScopeKind,
	}
}

func mapMember(d *ent.PrincipalRole) *RoleMember {
	return &RoleMember{
		PrincipalType: d.PrincipalType,
		PrincipalID:   d.PrincipalID,
		CreatedAt:     d.CreatedAt,
	}
}
