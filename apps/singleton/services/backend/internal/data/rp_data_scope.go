package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entdatascope "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/datascope"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
)

type dataScopeRP struct {
	RP
}

func NewDataScopeRP(logger log.Logger, store *Store) biz.DataScopeRP {
	return &dataScopeRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_data_scope")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *dataScopeRP) Create(ctx context.Context, scope *biz.DataScope) error {
	builder := rp.store.GetClient(ctx).DataScope.Create().
		SetRoleCode(*scope.RoleCode).
		SetScopeType(*scope.ScopeType).
		SetNillableScopeConfig(scope.ScopeConfig).
		SetNillableTargetResource(scope.TargetResource)
	if err := builder.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *dataScopeRP) Update(ctx context.Context, fieldsMask []string, scope *biz.DataScope) error {
	updater := rp.store.GetClient(ctx).DataScope.UpdateOneID(*scope.ID)
	utils.Handler{
		"scope_type": {
			Condition: scope.ScopeType != nil,
			OnTrue:    func() { updater.SetScopeType(*scope.ScopeType) },
			OnFalse:   func() {},
		},
		"scope_config": {
			Condition: scope.ScopeConfig != nil,
			OnTrue:    func() { updater.SetScopeConfig(*scope.ScopeConfig) },
			OnFalse:   func() { updater.ClearScopeConfig() },
		},
		"target_resource": {
			Condition: scope.TargetResource != nil,
			OnTrue:    func() { updater.SetTargetResource(*scope.TargetResource) },
			OnFalse:   func() { updater.ClearTargetResource() },
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *dataScopeRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).DataScope.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *dataScopeRP) Get(ctx context.Context, id string) (*biz.DataScope, error) {
	d, err := rp.store.GetClient(ctx).DataScope.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapDataScope(d), nil
}

func (rp *dataScopeRP) Query(ctx context.Context, in *biz.DataScopeQueryIn) (*biz.DataScopeQueryOut, error) {
	query := rp.store.GetClient(ctx).DataScope.Query()
	entutil.WherePtr(query, in.RoleCode, entdatascope.RoleCodeEQ)
	entutil.WherePtr(query, in.ScopeType, entdatascope.ScopeTypeEQ)
	entutil.WherePtr(query, in.TargetResource, entdatascope.TargetResourceContainsFold)
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entdatascope.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entdatascope.FieldUpdatedAt)) },
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
	return &biz.DataScopeQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapDataScope),
	}, nil
}

func (rp *dataScopeRP) GetScopesForRoles(ctx context.Context, roleCodes []string) ([]*datascope.RoleScope, error) {
	if len(roleCodes) == 0 {
		return nil, nil
	}

	scopes, err := rp.store.GetClient(ctx).DataScope.Query().
		Where(entdatascope.RoleCodeIn(roleCodes...)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	result := make([]*datascope.RoleScope, len(scopes))
	for i, s := range scopes {
		result[i] = &datascope.RoleScope{
			RoleCode:  s.RoleCode,
			ScopeType: s.ScopeType,
			ScopeConfig: func() string {
				if s.ScopeConfig != nil {
					return *s.ScopeConfig
				}
				return ""
			}(),
			TargetResource: s.TargetResource,
		}
	}
	return result, nil
}

func (rp *dataScopeRP) DeleteByRoleCode(ctx context.Context, roleCode string) error {
	_, err := rp.store.GetClient(ctx).DataScope.Delete().
		Where(entdatascope.RoleCodeEQ(roleCode)).
		Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapDataScope(d *ent.DataScope) *biz.DataScope {
	return &biz.DataScope{
		ID:             &d.ID,
		CreatedAt:      &d.CreatedAt,
		UpdatedAt:      &d.UpdatedAt,
		RoleCode:       &d.RoleCode,
		ScopeType:      &d.ScopeType,
		ScopeConfig:    d.ScopeConfig,
		TargetResource: &d.TargetResource,
	}
}
