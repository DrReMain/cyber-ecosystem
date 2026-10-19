package dept

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/dept"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type deptRP struct {
	shared.RP
}

func NewDeptRP(logger *slog.Logger, p *platform.Platform) DeptRP {
	return &deptRP{
		RP: shared.NewRP(logger.With("module", "module/dept_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *deptRP) Create(ctx context.Context, d *Dept) (*Dept, error) {
	created, err := rp.Platform.GetClient(ctx).Dept.Create().
		SetNillableParentID(d.ParentID).
		SetName(*d.Name).
		SetNillableRemark(d.Remark).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapDept(created), nil
}

func (rp *deptRP) Update(ctx context.Context, fieldsMask []string, d *Dept) (*Dept, error) {
	updater := rp.Platform.GetClient(ctx).Dept.UpdateOneID(d.ID)
	helper.Handler{
		"name": {
			Condition: d.Name != nil,
			OnTrue:    func() { updater.SetName(*d.Name) },
			OnFalse:   func() {},
		},
		"parent_id": {
			Condition: d.ParentID != nil,
			OnTrue:    func() { updater.SetNillableParentID(d.ParentID) },
			OnFalse:   func() { updater.ClearParentID() },
		},
		"remark": {
			Condition: d.Remark != nil,
			OnTrue:    func() { updater.SetRemark(*d.Remark) },
			OnFalse:   func() { updater.SetRemark("") },
		},
	}.Emit(fieldsMask)
	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapDept(updated), nil
}

func (rp *deptRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.Platform.GetClient(ctx).Dept.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *deptRP) List(ctx context.Context, in *DeptListIn) (*DeptListOut, error) {
	query := rp.Platform.GetClient(ctx).Dept.Query()
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), dept.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), dept.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtA), dept.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtZ), dept.UpdatedAtLTE)
	helper.Where(query, in.Name != nil, func() predicate.Dept { return dept.NameContainsFold(*in.Name) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"name":      func(sel helper.SQLSelector) { query.Order(sel(dept.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(dept.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(dept.FieldUpdatedAt)) },
	})

	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	ds, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	var counts []struct {
		DeptID string `json:"dept_id"`
		Count  int    `json:"count"`
	}
	if err := rp.Platform.GetClient(ctx).User.Query().
		GroupBy(user.FieldDeptID).
		Aggregate(ent.Count()).
		Scan(ctx, &counts); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	byDept := make(map[string]int64, len(counts))
	for _, c := range counts {
		byDept[c.DeptID] = int64(c.Count)
	}
	list := utils.SliceMap(ds, mapDept)
	for _, d := range list {
		d.UserCount = byDept[d.ID]
	}
	return &DeptListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         list,
	}, nil
}

func (rp *deptRP) Get(ctx context.Context, id string) (*Dept, error) {
	d, err := rp.Platform.GetClient(ctx).Dept.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapDept(d), nil
}

func (rp *deptRP) HasChildren(ctx context.Context, id string) (bool, error) {
	count, err := rp.Platform.GetClient(ctx).Dept.Query().
		Where(dept.ParentIDEQ(id)).
		Count(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return count > 0, nil
}

func (rp *deptRP) HasUsers(ctx context.Context, id string) (bool, error) {
	count, err := rp.Platform.GetClient(ctx).User.Query().
		Where(user.DeptIDEQ(id)).
		Count(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return count > 0, nil
}

func (rp *deptRP) IsAncestor(ctx context.Context, ancestorID, descendantID string) (bool, error) {
	client := rp.Platform.GetClient(ctx)
	cur := descendantID
	visited := map[string]struct{}{cur: {}}
	for {
		d, err := client.Dept.Get(ctx, cur)
		if err != nil {
			return false, rp.Platform.HandleEntError(err)
		}
		if d.ParentID == nil {
			return false, nil // reached root without a match
		}
		if *d.ParentID == ancestorID {
			return true, nil
		}
		if _, seen := visited[*d.ParentID]; seen {
			return false, nil // cycle guard — data should never form a ring
		}
		visited[*d.ParentID] = struct{}{}
		cur = *d.ParentID
	}
}

func (rp *deptRP) RemoveMember(ctx context.Context, deptID, userID string) (bool, error) {
	n, err := rp.Platform.GetClient(ctx).User.Update().
		Where(user.ID(userID), user.DeptIDEQ(deptID)).
		ClearDeptID().
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return n > 0, nil
}

func (rp *deptRP) ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error) {
	query := rp.Platform.GetClient(ctx).User.Query().
		Where(user.DeptIDEQ(in.DeptID)).
		Order(ent.Asc(user.FieldCreatedAt))
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	us, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &MemberListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(us, mapMember),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapDept(d *ent.Dept) *Dept {
	return &Dept{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		Name:      &d.Name,
		ParentID:  d.ParentID,
		Remark:    &d.Remark,
	}
}

func mapMember(d *ent.User) *DeptMember {
	return &DeptMember{
		UserID:  d.ID,
		Email:   d.Email,
		Enabled: d.Enabled,
	}
}
