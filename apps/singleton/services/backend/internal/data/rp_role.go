package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entrole "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/role"
)

type roleRP struct {
	RP
}

func NewRoleRP(logger log.Logger, store *Store) biz.RoleRP {
	return &roleRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_role")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *roleRP) Create(ctx context.Context, role *biz.Role) error {
	if err := rp.store.GetClient(ctx).Role.Create().
		SetName(*role.Name).
		SetCode(*role.Code).
		SetNillableDescription(role.Description).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *roleRP) Update(ctx context.Context, fieldsMask []string, role *biz.Role) error {
	updater := rp.store.GetClient(ctx).Role.UpdateOneID(*role.ID)
	utils.Handler{
		"name": {
			Condition: role.Name != nil,
			OnTrue:    func() { updater.SetName(*role.Name) },
			OnFalse:   func() {},
		},
		"description": {
			Condition: role.Description != nil,
			OnTrue:    func() { updater.SetDescription(*role.Description) },
			OnFalse:   func() { updater.SetDescription("") },
		},
		"status": {
			Condition: role.Status != nil,
			OnTrue:    func() { updater.SetStatus(*role.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *roleRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).Role.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *roleRP) Get(ctx context.Context, id string) (*biz.Role, error) {
	d, err := rp.store.GetClient(ctx).Role.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapRole(d), nil
}

func (rp *roleRP) Query(ctx context.Context, in *biz.RoleQueryIn) (*biz.RoleQueryOut, error) {
	query := rp.store.GetClient(ctx).Role.Query()
	entutil.WherePtr(query, in.Name, entrole.NameContainsFold)
	entutil.WherePtr(query, in.Code, entrole.CodeEQ)
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entrole.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entrole.FieldUpdatedAt)) },
		"sort":       func(sel entutil.SQLSelector) { query.Order(sel(entrole.FieldSort)) },
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
	return &biz.RoleQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapRole),
	}, nil
}

func (rp *roleRP) FindByCodes(ctx context.Context, codes []string) ([]*biz.Role, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	pos, err := rp.store.GetClient(ctx).Role.Query().
		Where(entrole.CodeIn(codes...)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return utils.SliceMap(pos, mapRole), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapRole(d *ent.Role) *biz.Role {
	return &biz.Role{
		ID:          &d.ID,
		CreatedAt:   &d.CreatedAt,
		UpdatedAt:   &d.UpdatedAt,
		Name:        &d.Name,
		Code:        &d.Code,
		Description: &d.Description,
		Status:      &d.Status,
		Sort:        &d.Sort,
	}
}
