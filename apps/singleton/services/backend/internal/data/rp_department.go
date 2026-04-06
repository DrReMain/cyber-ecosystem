package data

import (
	"context"

	"github.com/rs/xid"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entdepartment "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/department"
)

type departmentRP struct {
	RP
}

func NewDepartmentRP(logger log.Logger, store *Store) biz.DepartmentRP {
	return &departmentRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_department")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *departmentRP) Create(ctx context.Context, dept *biz.Department) error {
	id := xid.New().String()

	var path string
	if dept.ParentID != nil {
		parent, err := rp.store.GetClient(ctx).Department.Get(ctx, *dept.ParentID)
		if err != nil {
			return HandleError(err)
		}
		path = parent.Path + id + "/"
	} else {
		path = "/" + id + "/"
	}

	_, err := rp.store.GetClient(ctx).Department.Create().
		SetID(id).
		SetName(*dept.Name).
		SetCode(*dept.Code).
		SetPath(path).
		SetNillableParentID(dept.ParentID).
		Save(ctx)
	return HandleError(err)
}

func (rp *departmentRP) Update(ctx context.Context, fieldsMask []string, dept *biz.Department) error {
	updater := rp.store.GetClient(ctx).Department.UpdateOneID(*dept.ID)
	utils.Handler{
		"name": {
			Condition: dept.Name != nil,
			OnTrue:    func() { updater.SetName(*dept.Name) },
			OnFalse:   func() {},
		},
		"code": {
			Condition: dept.Code != nil,
			OnTrue:    func() { updater.SetCode(*dept.Code) },
			OnFalse:   func() {},
		},
		"status": {
			Condition: dept.Status != nil,
			OnTrue:    func() { updater.SetStatus(*dept.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *departmentRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).Department.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *departmentRP) Get(ctx context.Context, id string) (*biz.Department, error) {
	d, err := rp.store.GetClient(ctx).Department.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapDepartment(d), nil
}

func (rp *departmentRP) Query(ctx context.Context, in *biz.DepartmentQueryIn) (*biz.DepartmentQueryOut, error) {
	query := rp.store.GetClient(ctx).Department.Query()
	entutil.WherePtr(query, in.Name, entdepartment.NameContainsFold)
	entutil.WherePtr(query, in.Code, entdepartment.CodeEQ)
	entutil.WherePtr(query, in.ParentID, entdepartment.ParentIDEQ)
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entdepartment.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entdepartment.FieldUpdatedAt)) },
		"sort":       func(sel entutil.SQLSelector) { query.Order(sel(entdepartment.FieldSort)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		singletonV1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}
	pos, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return &biz.DepartmentQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapDepartment),
	}, nil
}

func (rp *departmentRP) GetDescendantDeptIDs(ctx context.Context, parentIDs []string) ([]string, error) {
	if len(parentIDs) == 0 {
		return nil, nil
	}
	client := rp.store.GetClient(ctx)
	seen := make(map[string]struct{}, len(parentIDs)*3)

	for _, pid := range parentIDs {
		dept, err := client.Department.Get(ctx, pid)
		if err != nil {
			return nil, HandleError(err)
		}
		seen[pid] = struct{}{}
		descendantIDs, err := client.Department.Query().
			Where(
				entdepartment.PathHasPrefix(dept.Path),
				entdepartment.IDNEQ(pid),
			).
			IDs(ctx)
		if err != nil {
			return nil, HandleError(err)
		}
		for _, id := range descendantIDs {
			seen[id] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result, nil
}

func (rp *departmentRP) HasChildren(ctx context.Context, parentID string) (bool, error) {
	count, err := rp.store.GetClient(ctx).Department.Query().
		Where(entdepartment.ParentIDEQ(parentID)).
		Limit(1).
		Count(ctx)
	if err != nil {
		return false, HandleError(err)
	}
	return count > 0, nil
}

func (rp *departmentRP) Move(ctx context.Context, id string, targetParent *biz.Department) error {
	client := rp.store.GetClient(ctx)
	dept, err := client.Department.Get(ctx, id)
	if err != nil {
		return HandleError(err)
	}

	oldPath := dept.Path
	var newPath string
	updater := client.Department.UpdateOneID(id)

	if targetParent != nil {
		newPath = *targetParent.Path + dept.ID + "/"
		updater.SetPath(newPath).SetParentID(*targetParent.ID)
	} else {
		newPath = "/" + dept.ID + "/"
		updater.SetPath(newPath).ClearParentID()
	}

	// Bulk update all descendants: replace oldPath prefix with newPath
	pattern := oldPath + "%"
	_, err = client.ExecContext(ctx,
		"UPDATE department SET path = CONCAT($1::text, SUBSTRING(path, $2::integer)) WHERE path LIKE $3::text AND id != $4::text",
		newPath, len(oldPath)+1, pattern, id,
	)
	if err != nil {
		return HandleError(err)
	}

	// Update self (path + parent_id)
	if _, err := updater.Save(ctx); err != nil {
		return HandleError(err)
	}

	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapDepartment(d *ent.Department) *biz.Department {
	return &biz.Department{
		ID:        &d.ID,
		CreatedAt: &d.CreatedAt,
		UpdatedAt: &d.UpdatedAt,
		Name:      &d.Name,
		Code:      &d.Code,
		ParentID:  d.ParentID,
		Path:      &d.Path,
		Status:    &d.Status,
		Sort:      &d.Sort,
	}
}
