package dept

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/dept"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

type deptRP struct {
	shared.RP
}

func NewDeptRP(logger *slog.Logger, p *platform.Platform) DeptRP {
	return &deptRP{
		RP: shared.NewRP(logger.With("module", "module/dept_rp"), p),
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *deptRP) Create(ctx context.Context, d *Dept) (*Dept, error) {
	created, err := rp.Platform.GetClient(ctx).Dept.Create().
		SetTenantID(d.TenantID).
		SetNillableParentID(d.ParentID).
		SetName(*d.Name).
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

	total, offset, limit, err := helper.ApplyPagination(ctx, query, in.PageRequest,
		helper.NewPageConfig(helper.DefaultPageSize, helper.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	ds, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &DeptListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(ds, mapDept),
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

// IsAncestor reports whether ancestorID is an ancestor of descendantID by walking the
// parent chain upward from descendantID. O(depth); a visited guard breaks any
// (defensively unexpected) cycle so the loop can never spin forever.
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

// Private -------------------------------------------------------------------------------------------------------------

func mapDept(d *ent.Dept) *Dept {
	return &Dept{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		Name:      &d.Name,
		ParentID:  d.ParentID,
	}
}
