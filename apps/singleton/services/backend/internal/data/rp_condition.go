package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entcondition "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/condition"
)

type conditionRP struct {
	RP
}

func NewConditionRP(logger log.Logger, store *Store) biz.ConditionRP {
	return &conditionRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_condition")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *conditionRP) Create(ctx context.Context, cond *biz.Condition) error {
	builder := rp.store.GetClient(ctx).Condition.Create().
		SetRoleCode(*cond.RoleCode).
		SetOperation(*cond.Operation).
		SetConditionType(*cond.ConditionType).
		SetNillableConfig(cond.Config).
		SetNillableGroupID(cond.GroupID)
	if err := builder.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *conditionRP) Update(ctx context.Context, fieldsMask []string, cond *biz.Condition) error {
	updater := rp.store.GetClient(ctx).Condition.UpdateOneID(*cond.ID)
	utils.Handler{
		"role_code": {
			Condition: cond.RoleCode != nil,
			OnTrue:    func() { updater.SetRoleCode(*cond.RoleCode) },
			OnFalse:   func() {},
		},
		"operation": {
			Condition: cond.Operation != nil,
			OnTrue:    func() { updater.SetOperation(*cond.Operation) },
			OnFalse:   func() {},
		},
		"condition_type": {
			Condition: cond.ConditionType != nil,
			OnTrue:    func() { updater.SetConditionType(*cond.ConditionType) },
			OnFalse:   func() {},
		},
		"config": {
			Condition: cond.Config != nil,
			OnTrue:    func() { updater.SetConfig(*cond.Config) },
			OnFalse:   func() { updater.ClearConfig() },
		},
		"group_id": {
			Condition: cond.GroupID != nil,
			OnTrue:    func() { updater.SetGroupID(*cond.GroupID) },
			OnFalse:   func() { updater.ClearGroupID() },
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *conditionRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).Condition.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *conditionRP) Get(ctx context.Context, id string) (*biz.Condition, error) {
	d, err := rp.store.GetClient(ctx).Condition.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapCondition(d), nil
}

func (rp *conditionRP) Query(ctx context.Context, in *biz.ConditionQueryIn) (*biz.ConditionQueryOut, error) {
	query := rp.store.GetClient(ctx).Condition.Query()
	entutil.WherePtr(query, in.RoleCode, entcondition.RoleCodeEQ)
	entutil.WherePtr(query, in.Operation, entcondition.OperationContainsFold)
	entutil.WherePtr(query, in.ConditionType, entcondition.ConditionTypeEQ)
	entutil.WherePtr(query, in.GroupID, entcondition.GroupIDEQ)
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entcondition.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entcondition.FieldUpdatedAt)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		singletonV1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return &biz.ConditionQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(items, mapCondition),
	}, nil
}

func (rp *conditionRP) DeleteByRoleCode(ctx context.Context, roleCode string) error {
	_, err := rp.store.GetClient(ctx).Condition.Delete().
		Where(entcondition.RoleCodeEQ(roleCode)).
		Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *conditionRP) ListByRoleCodes(ctx context.Context, roleCodes []string) ([]*biz.Condition, error) {
	if len(roleCodes) == 0 {
		return nil, nil
	}
	items, err := rp.store.GetClient(ctx).Condition.Query().
		Where(entcondition.RoleCodeIn(roleCodes...)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return utils.SliceMap(items, mapCondition), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapCondition(d *ent.Condition) *biz.Condition {
	cond := &biz.Condition{
		ID:            &d.ID,
		CreatedAt:     &d.CreatedAt,
		UpdatedAt:     &d.UpdatedAt,
		RoleCode:      &d.RoleCode,
		Operation:     &d.Operation,
		ConditionType: &d.ConditionType,
		Config:        d.Config,
		GroupID:       d.GroupID,
	}
	return cond
}
