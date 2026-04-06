package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/condition"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

// Model ---------------------------------------------------------------------------------------------------------------

// Condition represents an ABAC condition rule bound to a role.
type Condition struct {
	ID            *string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	RoleCode      *string
	Operation     *string
	ConditionType *string
	Config        *string
	GroupID       *string
}

// ConditionQueryIn is the query input for access conditions.
type ConditionQueryIn struct {
	*common.PageRequest
	OrderBy       []*utils.OrderBy
	RoleCode      *string
	Operation     *string
	ConditionType *string
	GroupID       *string
}

// ConditionQueryOut is the query output for access conditions.
type ConditionQueryOut struct {
	*common.PageResponse
	List []*Condition
}

// Port ----------------------------------------------------------------------------------------------------------------

// ConditionRP abstracts access condition persistence.
type ConditionRP interface {
	Create(ctx context.Context, cond *Condition) error
	Update(ctx context.Context, fieldsMask []string, cond *Condition) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Condition, error)
	Query(ctx context.Context, in *ConditionQueryIn) (*ConditionQueryOut, error)
	ListByRoleCodes(ctx context.Context, roleCodes []string) ([]*Condition, error)
	DeleteByRoleCode(ctx context.Context, roleCode string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

// ConditionUC handles ABAC condition operations: CRUD, condition evaluation.
type ConditionUC struct {
	UC
	plugins  *condition.ConditionRegistry
	policyUC *PolicyUC
	condRP   ConditionRP
}

// NewConditionUC creates a new ConditionUC.
func NewConditionUC(logger log.Logger, tm Transaction, plugins *condition.ConditionRegistry, policyUC *PolicyUC, condRP ConditionRP) *ConditionUC {
	return &ConditionUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_condition")),
			tm:  tm,
		},
		plugins:  plugins,
		policyUC: policyUC,
		condRP:   condRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

// CheckConditions checks if the user satisfies conditions for the given operation.
// Evaluation is per-role: for each role, conditions are grouped by group_id (AND within group, OR between groups).
// If any role passes (or has no conditions for the operation), the request is allowed.
func (uc *ConditionUC) CheckConditions(ctx context.Context, userID, operation string) (bool, error) {
	roles := uc.policyUC.GetRolesForUser(userID)
	if len(roles) == 0 {
		return true, nil
	}

	conditions, err := uc.condRP.ListByRoleCodes(ctx, roles)
	if err != nil {
		return false, err
	}

	roleConds := groupByRole(conditions, operation)
	if len(roleConds) == 0 {
		return true, nil
	}

	for _, role := range roles {
		conds, ok := roleConds[role]
		if !ok {
			return true, nil
		}
		if passes, err := uc.evaluateRole(ctx, conds); err != nil {
			return false, err
		} else if passes {
			return true, nil
		}
	}
	return false, nil
}

// evaluateRole evaluates all conditions for a single role.
// Conditions are grouped by group_id: AND within group, OR between groups.
func (uc *ConditionUC) evaluateRole(ctx context.Context, conds []*Condition) (bool, error) {
	groups := make(map[string][]*Condition)
	for _, cond := range conds {
		gid := utils.Deref(cond.GroupID, "")
		groups[gid] = append(groups[gid], cond)
	}
	for _, group := range groups {
		if passes, err := uc.evaluateGroup(ctx, group); err != nil {
			return false, err
		} else if passes {
			return true, nil
		}
	}
	return false, nil
}

// evaluateGroup evaluates all conditions in a group with AND semantics.
func (uc *ConditionUC) evaluateGroup(ctx context.Context, conds []*Condition) (bool, error) {
	for _, cond := range conds {
		ct := utils.Deref(cond.ConditionType, "")
		config := utils.Deref(cond.Config, "")
		allowed, err := uc.plugins.Evaluate(ctx, ct, config)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// CRUD -----------------------------------------------------------------------------------------------------------------

func (uc *ConditionUC) validateCondition(cond *Condition) error {
	ct := ""
	if cond.ConditionType != nil {
		ct = *cond.ConditionType
	}
	if ct == "" {
		return singletonV1.ErrorErrorReasonInvalidArgument("condition_type is required")
	}
	if _, ok := uc.plugins.Get(ct); !ok {
		return singletonV1.ErrorErrorReasonInvalidArgument("unsupported condition_type: %s", ct)
	}
	config := ""
	if cond.Config != nil {
		config = *cond.Config
	}
	if config == "" {
		return nil
	}
	if err := uc.plugins.Validate(ct, config); err != nil {
		return singletonV1.ErrorErrorReasonInvalidArgument("%s", err.Error())
	}
	return nil
}

// CreateCondition creates a new access condition.
func (uc *ConditionUC) CreateCondition(ctx context.Context, cond *Condition) error {
	if err := uc.validateCondition(cond); err != nil {
		return err
	}
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		return uc.condRP.Create(txCtx, cond)
	})
}

// UpdateCondition updates an access condition.
func (uc *ConditionUC) UpdateCondition(ctx context.Context, fieldsMask []string, cond *Condition) error {
	if err := uc.validateCondition(cond); err != nil {
		return err
	}
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		return uc.condRP.Update(txCtx, fieldsMask, cond)
	})
}

// DeleteCondition deletes an access condition.
func (uc *ConditionUC) DeleteCondition(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		return uc.condRP.Delete(txCtx, id)
	})
}

// GetCondition returns an access condition by ID.
func (uc *ConditionUC) GetCondition(ctx context.Context, id string) (*Condition, error) {
	return uc.condRP.Get(ctx, id)
}

// QueryConditions queries access conditions with pagination.
func (uc *ConditionUC) QueryConditions(ctx context.Context, in *ConditionQueryIn) (*ConditionQueryOut, error) {
	return uc.condRP.Query(ctx, in)
}

// CascadeRoleDelete removes all access conditions for a role.
func (uc *ConditionUC) CascadeRoleDelete(ctx context.Context, roleCode string) error {
	return uc.condRP.DeleteByRoleCode(ctx, roleCode)
}

// ---------------------------------------------------------------------------------------------------------------------

func groupByRole(conditions []*Condition, operation string) map[string][]*Condition {
	result := make(map[string][]*Condition)
	for _, cond := range conditions {
		op := utils.Deref(cond.Operation, "")
		if !security.MatchResource(op, operation) {
			continue
		}
		rc := utils.Deref(cond.RoleCode, "")
		result[rc] = append(result[rc], cond)
	}
	return result
}
