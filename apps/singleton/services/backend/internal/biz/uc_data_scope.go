package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
)

// Model ---------------------------------------------------------------------------------------------------------------

// DataScope represents a data scope rule bound to a role.
type DataScope struct {
	ID             *string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	RoleCode       *string
	ScopeType      *string
	ScopeConfig    *string
	TargetResource *string
}

// DataScopeQueryIn is the query input for data scopes.
type DataScopeQueryIn struct {
	*common.PageRequest
	OrderBy        []*utils.OrderBy
	RoleCode       *string
	ScopeType      *string
	TargetResource *string
}

// DataScopeQueryOut is the query output for data scopes.
type DataScopeQueryOut struct {
	*common.PageResponse
	List []*DataScope
}

// Port ----------------------------------------------------------------------------------------------------------------

// DataScopeRP abstracts data scope persistence.
type DataScopeRP interface {
	Create(ctx context.Context, scope *DataScope) error
	Update(ctx context.Context, fieldsMask []string, scope *DataScope) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*DataScope, error)
	Query(ctx context.Context, in *DataScopeQueryIn) (*DataScopeQueryOut, error)
	GetScopesForRoles(ctx context.Context, roleCodes []string) ([]*datascope.RoleScope, error)
	DeleteByRoleCode(ctx context.Context, roleCode string) error
}

// DataScopeSnapshotRP abstracts scope snapshot caching.
type DataScopeSnapshotRP interface {
	Get(ctx context.Context, userID string) (*datascope.ScopeSnapshot, bool)
	Set(ctx context.Context, userID string, snapshot *datascope.ScopeSnapshot) error
	Invalidate(ctx context.Context, userID string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

// DataScopeUC handles data permission operations: scope resolution, scope CRUD, snapshot caching.
type DataScopeUC struct {
	UC
	policyUC      *PolicyUC
	dataScopeRP   DataScopeRP
	snapshotRP    DataScopeSnapshotRP
	deptBindingRP DepartmentBindingRP
	deptRP        DepartmentRP
	userAttrRP    UserAttributeRP
	scopePlugins  *datascope.ScopePluginRegistry
}

// NewDataScopeUC creates a new DataScopeUC.
func NewDataScopeUC(
	logger log.Logger,
	tm Transaction,
	policyUC *PolicyUC,
	dataScopeRP DataScopeRP,
	snapshotRP DataScopeSnapshotRP,
	deptBindingRP DepartmentBindingRP,
	deptRP DepartmentRP,
	userAttrRP UserAttributeRP,
	scopePlugins *datascope.ScopePluginRegistry,
) *DataScopeUC {
	return &DataScopeUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_data_scope")),
			tm:  tm,
		},
		policyUC:      policyUC,
		dataScopeRP:   dataScopeRP,
		snapshotRP:    snapshotRP,
		deptBindingRP: deptBindingRP,
		deptRP:        deptRP,
		userAttrRP:    userAttrRP,
		scopePlugins:  scopePlugins,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

// ResolveScope resolves the effective data scope for a user and operation.
func (uc *DataScopeUC) ResolveScope(ctx context.Context, userID, operation string) (*datascope.EffectiveScope, error) {
	snapshot, err := uc.buildSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}

	matched := matchScopes(snapshot.Scopes, operation)

	return mergeScopes(matched, snapshot, uc.scopePlugins)
}

// GetDataScope returns a single data scope by ID.
func (uc *DataScopeUC) GetDataScope(ctx context.Context, id string) (*DataScope, error) {
	return uc.dataScopeRP.Get(ctx, id)
}

// QueryDataScopes queries data scopes with pagination.
func (uc *DataScopeUC) QueryDataScopes(ctx context.Context, in *DataScopeQueryIn) (*DataScopeQueryOut, error) {
	return uc.dataScopeRP.Query(ctx, in)
}

// CreateDataScope creates a new data scope and invalidates the affected role cache.
func (uc *DataScopeUC) CreateDataScope(ctx context.Context, scope *DataScope) error {
	if err := uc.validateScope(scope); err != nil {
		return err
	}
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		if err := uc.dataScopeRP.Create(txCtx, scope); err != nil {
			return err
		}
		return uc.InvalidateRole(txCtx, *scope.RoleCode)
	})
}

// UpdateDataScope updates a data scope and invalidates the affected role cache.
func (uc *DataScopeUC) UpdateDataScope(ctx context.Context, fieldsMask []string, scope *DataScope) error {
	if err := uc.validateScope(scope); err != nil {
		return err
	}
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		existing, err := uc.dataScopeRP.Get(txCtx, *scope.ID)
		if err != nil {
			return err
		}
		if err := uc.dataScopeRP.Update(txCtx, fieldsMask, scope); err != nil {
			return err
		}
		return uc.InvalidateRole(txCtx, *existing.RoleCode)
	})
}

// DeleteDataScope deletes a data scope and invalidates the affected role cache.
func (uc *DataScopeUC) DeleteDataScope(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		scope, err := uc.dataScopeRP.Get(txCtx, id)
		if err != nil {
			return err
		}
		if err := uc.dataScopeRP.Delete(txCtx, id); err != nil {
			return err
		}
		return uc.InvalidateRole(txCtx, *scope.RoleCode)
	})
}

// InvalidateUser clears the cached scope snapshot for a single user.
func (uc *DataScopeUC) InvalidateUser(ctx context.Context, userID string) error {
	return uc.snapshotRP.Invalidate(ctx, userID)
}

// InvalidateRole clears cached scope snapshots for all users holding the given role.
func (uc *DataScopeUC) InvalidateRole(ctx context.Context, roleCode string) error {
	users := uc.policyUC.GetUsersForRole(roleCode)
	for _, userID := range users {
		if err := uc.snapshotRP.Invalidate(ctx, userID); err != nil {
			uc.log.Warnf("failed to invalidate scope for user %s: %v", userID, err)
		}
	}
	return nil
}

// CascadeRoleDelete removes all data scopes and cached snapshots for a role.
func (uc *DataScopeUC) CascadeRoleDelete(ctx context.Context, roleCode string) error {
	if err := uc.InvalidateRole(ctx, roleCode); err != nil {
		return err
	}
	return uc.dataScopeRP.DeleteByRoleCode(ctx, roleCode)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *DataScopeUC) validateScope(scope *DataScope) error {
	st := ""
	if scope.ScopeType != nil {
		st = *scope.ScopeType
	}
	if st == "" {
		return singletonV1.ErrorErrorReasonInvalidArgument("scope_type is required")
	}
	if _, ok := uc.scopePlugins.Get(st); !ok {
		return singletonV1.ErrorErrorReasonInvalidArgument("unsupported scope_type: %s", st)
	}
	config := ""
	if scope.ScopeConfig != nil {
		config = *scope.ScopeConfig
	}
	if err := uc.scopePlugins.Validate(st, config); err != nil {
		return singletonV1.ErrorErrorReasonInvalidArgument("%s", err.Error())
	}
	return nil
}

// buildSnapshot builds (or retrieves from cache) a scope snapshot for the given user.
func (uc *DataScopeUC) buildSnapshot(ctx context.Context, userID string) (*datascope.ScopeSnapshot, error) {
	if snapshot, ok := uc.snapshotRP.Get(ctx, userID); ok {
		return snapshot, nil
	}

	snapshot := &datascope.ScopeSnapshot{}

	roles := uc.policyUC.GetRolesForUser(userID)
	if len(roles) == 0 {
		return &datascope.ScopeSnapshot{}, nil
	}
	snapshot.Roles = roles

	scopes, err := uc.dataScopeRP.GetScopesForRoles(ctx, roles)
	if err != nil {
		return nil, err
	}
	scopeList := make([]datascope.RoleScope, len(scopes))
	for i, s := range scopes {
		scopeList[i] = *s
	}
	snapshot.Scopes = scopeList

	hasDept := false
	hasAttribute := false
	for _, s := range snapshot.Scopes {
		if s.ScopeType == "dept" {
			hasDept = true
		}
		if s.ScopeType == "attribute" {
			hasAttribute = true
		}
	}

	if hasDept {
		deptList, err := uc.deptBindingRP.ListUserDepartments(ctx, userID)
		if err != nil {
			return nil, err
		}
		directIDs := make([]string, 0, len(deptList))
		for _, dept := range deptList {
			if dept.ID != nil {
				directIDs = append(directIDs, *dept.ID)
			}
		}
		allDeptIDs, err := uc.deptRP.GetDescendantDeptIDs(ctx, directIDs)
		if err != nil {
			return nil, err
		}
		snapshot.DeptIDs = allDeptIDs
	}

	if hasAttribute {
		attrs, err := uc.userAttrRP.Query(ctx, userID)
		if err != nil {
			return nil, err
		}
		attrMap := make(map[string]string, len(attrs))
		for _, a := range attrs {
			if a.Key != nil && a.Value != nil {
				attrMap[*a.Key] = *a.Value
			}
		}
		snapshot.Attributes = attrMap
	}

	snapshot.CachedAt = time.Now()

	if err := uc.snapshotRP.Set(ctx, userID, snapshot); err != nil {
		uc.log.Warnf("failed to cache scope snapshot: %v", err)
	}

	return snapshot, nil
}

// matchScopes filters scopes whose TargetResource matches the given operation.
func matchScopes(scopes []datascope.RoleScope, operation string) []datascope.RoleScope {
	var matched []datascope.RoleScope
	for _, s := range scopes {
		if datascope.MatchResource(s.TargetResource, operation) {
			matched = append(matched, s)
		}
	}
	return matched
}

// mergeScopes merges matched scopes into a single EffectiveScope using the plugin registry.
func mergeScopes(matched []datascope.RoleScope, snap *datascope.ScopeSnapshot, plugins *datascope.ScopePluginRegistry) (*datascope.EffectiveScope, error) {
	if len(matched) == 0 {
		return &datascope.EffectiveScope{IsAll: false}, nil
	}

	result := &datascope.EffectiveScope{}
	for _, s := range matched {
		if err := plugins.Merge(s, snap, result); err != nil {
			return nil, err
		}
		if result.IsAll {
			return result, nil
		}
	}

	result.DeptIDs = datascope.UniqueStrings(result.DeptIDs)

	if !result.SelfFilter && !result.DeptFilter && !result.AttributeFilter && len(result.ExtraPredicates) == 0 {
		result.SelfFilter = true
	}

	return result, nil
}
